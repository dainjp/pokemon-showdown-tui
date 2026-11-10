package tui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/coder/websocket"
	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
	"github.com/unnipv/pokemon-slowdown/internal/teams"
)

func challengeTestClient(t *testing.T) (*showdown.Client, <-chan string) {
	t.Helper()
	commands := make(chan string, 16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		for {
			_, data, err := conn.Read(context.Background())
			if err != nil {
				return
			}
			commands <- string(data)
		}
	}))
	t.Cleanup(server.Close)
	client := showdown.NewClient("ws" + strings.TrimPrefix(server.URL, "http"))
	client.Start()
	t.Cleanup(client.Close)
	select {
	case event := <-client.Events():
		if _, ok := event.(showdown.Connected); !ok {
			t.Fatalf("mock connection failed: %T", event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("mock connection timed out")
	}
	return client, commands
}

func expectBattleCommand(t *testing.T, commands <-chan string, want string) {
	t.Helper()
	select {
	case got := <-commands:
		if got != want {
			t.Fatalf("command = %q, want %q", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("no command; want %q", want)
	}
}

func challengeTestTeams() *teams.Store {
	return &teams.Store{Teams: []teams.Team{
		{Name: "Other format", Format: "gen8ou", Pokemon: []teams.Pokemon{{Species: "Bulbasaur", Moves: []string{"Tackle"}}}},
		{Name: "OU team", Format: "gen9ou", Pokemon: []teams.Pokemon{{Species: "Pikachu", Moves: []string{"Thunderbolt"}}}},
	}}
}

func TestBattleRequestsUploadCorrectTeamBeforeCommand(t *testing.T) {
	for _, format := range []string{"gen9ou", "gen9randombattle", "gen9battlefactory"} {
		for _, action := range []string{"challenge", "accept", "search"} {
			t.Run(action+"/"+format, func(t *testing.T) {
				client, commands := challengeTestClient(t)
				store := challengeTestTeams()
				m := New(config.Default(), Deps{Client: client, Teams: store})
				m.formats = []showdown.Format{{ID: format, Random: format != "gen9ou", Searchable: true, Challengeable: true}}
				f := m.formatInfo(format)
				var cmd tea.Cmd
				want := ""
				switch action {
				case "challenge":
					cmd = m.challengeFormat("Rival", f)
					want = "|/challenge Rival, " + format
				case "accept":
					cmd = m.acceptChallenge("Rival", format)
					want = "|/accept Rival"
				case "search":
					cmd = m.queueFormat(f)
					want = "|/search " + format
				}
				if cmd == nil {
					t.Fatal("request unexpectedly blocked")
				}
				if extra := m.queueFormat(f); extra != nil {
					t.Fatal("a second action could overwrite the first action's team")
				}
				m.Update(cmd())
				packed := "null"
				if format == "gen9ou" {
					packed = store.Teams[1].Pack()
				}
				expectBattleCommand(t, commands, "|/utm "+packed)
				expectBattleCommand(t, commands, want)
				if m.battleRequestPending {
					t.Fatal("completed request left controls blocked")
				}
			})
		}
	}
}

func TestMissingTeamBlocksChallengeAcceptanceAndSearch(t *testing.T) {
	for _, store := range []*teams.Store{nil, {}, {Teams: []teams.Team{{Name: "Empty"}}}} {
		m := New(config.Default(), Deps{Client: showdown.NewClient(""), Teams: store})
		f := m.formatInfo("gen9ou")
		for _, action := range []func() tea.Cmd{
			func() tea.Cmd { return m.challengeFormat("Rival", f) },
			func() tea.Cmd { return m.acceptChallenge("Rival", f.ID) },
			func() tea.Cmd { return m.queueFormat(f) },
		} {
			if cmd := action(); cmd != nil || m.battleRequestPending {
				t.Fatal("team-based request proceeded without a team")
			}
			if !strings.Contains(m.toast, "팀 가져오기가 필요합니다") {
				t.Fatalf("missing import guidance: %q", m.toast)
			}
		}
	}
}

func TestFailedTeamUploadStopsTheBattleCommand(t *testing.T) {
	m := New(config.Default(), Deps{Client: showdown.NewClient("")})
	sent := false
	cmd := m.withTeam(showdown.Format{Random: true}, func() error { sent = true; return nil })
	m.Update(cmd())
	if sent || m.battleRequestPending || !strings.Contains(m.toast, "대전 요청을 보내지 못했습니다") {
		t.Fatalf("failed upload was ignored: sent=%v, pending=%v, toast=%q", sent, m.battleRequestPending, m.toast)
	}
}

func TestChallengePickerUsesSelectedFormatAndEscapeCancels(t *testing.T) {
	client, commands := challengeTestClient(t)
	m := New(config.Default(), Deps{Client: client, Teams: challengeTestTeams()})
	m.formats = []showdown.Format{
		{ID: "ladderonly", Searchable: true},
		{ID: "gen9randombattle", Random: true, Searchable: true, Challengeable: true},
		{ID: "gen9ou", Searchable: true, Challengeable: true},
		{ID: "gen9customgame", Challengeable: true},
	}
	if cmd := m.submitChallenge("Rival"); cmd != nil {
		t.Fatal("username prompt sent a challenge before format selection")
	}
	if len(m.filteredFormats()) != 3 {
		t.Fatal("picker did not exclude ladder-only formats")
	}
	m.handleLobbyKey("end")
	m.handleLobbyKey("esc")
	if m.challengeUser != "" {
		t.Fatal("Escape did not leave challenge selection")
	}
	m.submitChallenge("Rival")
	m.formatFilter = "gen9ou"
	m.formatCursor = 0
	cmd := m.handleLobbyKey("enter")
	if cmd == nil || m.challengeUser != "" {
		t.Fatal("selected format was not submitted")
	}
	m.Update(cmd())
	expectBattleCommand(t, commands, "|/utm "+m.deps.Teams.Teams[1].Pack())
	expectBattleCommand(t, commands, "|/challenge Rival, gen9ou")
}

func TestCurrentChallengeUpdatesExposeCancellationAndTrackMultipleUsers(t *testing.T) {
	client, commands := challengeTestClient(t)
	m := New(config.Default(), Deps{Client: client})
	m.username = "‽Local@!away"
	feed := func(line string) {
		for _, event := range showdown.Parse(showdown.Frame{Lines: []string{line}}) {
			m.handleShowdownEvent(event)
		}
	}
	feed("|pm|‽Local|+Rival|/challenge gen9ou|gen9ou|||")
	feed("|pm|‽Local| Other|/challenge gen9randombattle|gen9randombattle|||")
	feed("|pm|@Friend|‽Local|/challenge gen9ou|gen9ou|||")
	if len(m.challengesTo) != 2 || m.challengesFrom["Friend"] != "gen9ou" {
		t.Fatal("current-server PM updates were not tracked")
	}
	var cancel tea.Cmd
	for _, command := range m.commands() {
		if command.title == "보낸 도전 취소: Rival" {
			cancel = command.run(m)
		}
	}
	if cancel == nil {
		t.Fatal("no cancellation action in the palette")
	}
	m.Update(cancel())
	expectBattleCommand(t, commands, "|/cancelchallenge Rival")
	feed("|pm| Local| Rival|/challenge")
	feed("|pm| Friend| Local|/challenge")
	if len(m.challengesTo) != 1 || len(m.challengesFrom) != 0 || m.challengesTo["Other"] == "" {
		t.Fatal("cancellation removed the wrong user's challenge")
	}
	m.handleShowdownEvent(showdown.ChallengesUpdated{})
	feed("|pm| Friend| Local|/challenge gen9ou|gen9ou|||")
	if m.challengesFrom["Friend"] != "gen9ou" {
		t.Fatal("legacy empty updates broke subsequent current-server updates")
	}
}

func TestChallengeViewsFitTerminalCells(t *testing.T) {
	m := New(config.Default(), Deps{})
	m.username, m.conn = "Trainer", connOnline
	m.formats = []showdown.Format{
		{ID: "gen9randombattle", Random: true, Searchable: true, Challengeable: true},
		{ID: "gen9ou", Searchable: true, Challengeable: true},
		{ID: "gen9customgame", Challengeable: true},
	}
	m.challengesTo["Rival"] = "gen9ou"
	for _, width := range []int{32, 60, 120} {
		m.width, m.height = width, 24
		m.layout = LayoutForSize(width, m.height)
		m.submitChallenge("Rival")
		for _, palette := range []bool{false, true} {
			m.palette.open = palette
			out := m.render()
			if lipgloss.Height(out) > m.height {
				t.Fatalf("challenge view exceeds %dx%d", width, m.height)
			}
			for _, line := range strings.Split(out, "\n") {
				if lipgloss.Width(line) > width {
					t.Fatalf("challenge line exceeds %d cells: %q", width, line)
				}
			}
			if dir := os.Getenv("SLOWDOWN_PREVIEW_DIR"); dir != "" {
				name := "challenge-" + map[int]string{32: "compact", 60: "narrow", 120: "wide"}[width]
				if palette {
					name += "-menu"
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name+".ansi"), []byte(out), 0644); err != nil {
					t.Fatal(err)
				}
			}
		}
		m.palette.open = false
	}
}
