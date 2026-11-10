package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
	"github.com/unnipv/pokemon-slowdown/internal/teams"
)

func decisionBattle(m *Model, forced []bool) *battleView {
	bv := m.battleFor("battle-gen9doublesou-123")
	s := bv.state()
	s.Me = "p1"
	s.Request = &battle.Request{RqID: 7, ForceSwitch: forced,
		Active: []battle.PokemonMoveRequest{
			{Moves: []battle.MoveRequest{{Move: "Protect", Target: "self"}}},
			{Moves: []battle.MoveRequest{{Move: "Protect", Target: "self"}}},
		}, Side: battle.SideRequest{Pokemon: []battle.PokemonSwitchRequest{
			{Active: true, Condition: "100/100"}, {Active: true, Condition: "100/100"},
			{Condition: "100/100"}, {Condition: "100/100"},
		}},
	}
	s.AwaitingChoice = true
	return bv
}

func TestDoublesWaitForBothChoicesAndUndo(t *testing.T) {
	client, wire := challengeTestClient(t)
	m := New(englishTestConfig(), Deps{Client: client})
	bv := decisionBattle(m, nil)
	if cmd := bv.chooseMove(1, m); cmd != nil || bv.slot != 1 || bv.waiting || bv.sending {
		t.Fatal("first move submitted early")
	}
	cmd := bv.chooseMove(1, m)
	if cmd == nil || bv.waiting || !bv.sending {
		t.Fatal("missing in-flight state")
	}
	m.Update(cmd())
	expectBattleCommand(t, wire, bv.room+"|/choose move 1, move 1|7")
	if !bv.waiting || bv.sending {
		t.Fatal("write success not reflected")
	}
	cmd = bv.undo()
	if cmd == nil {
		t.Fatal("undo unavailable")
	}
	m.Update(cmd())
	expectBattleCommand(t, wire, bv.room+"|/undo")
	if bv.waiting || bv.sending {
		t.Fatal("undo did not restore choices")
	}
}

func TestForcedSwitchesAndPasses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		forced []bool
		bench  int
		want   string
	}{
		{"both", []bool{true, true}, 2, "switch 3, switch 4"},
		{"second", []bool{false, true}, 2, "pass, switch 3"},
		{"one reserve", []bool{true, true}, 1, "switch 3, pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, wire := challengeTestClient(t)
			m := New(englishTestConfig(), Deps{Client: client})
			bv := decisionBattle(m, tc.forced)
			bv.state().Request.Side.Pokemon = bv.state().Request.Side.Pokemon[:2+tc.bench]
			bv.advanceSlot()
			cmd := bv.commitSwitch(3)
			if tc.name == "both" {
				if cmd != nil || bv.slot != 1 {
					t.Fatal("switch sent early")
				}
				if bv.commitSwitch(3) != nil || len(bv.draft) != 1 {
					t.Fatal("duplicate switch accepted")
				}
				cmd = bv.commitSwitch(4)
			}
			if cmd == nil {
				t.Fatal("complete switch was not submitted")
			}
			m.Update(cmd())
			expectBattleCommand(t, wire, bv.room+"|/choose "+tc.want+"|7")
		})
	}
}

func TestFaintedMoveSlotAndSendFailure(t *testing.T) {
	client := showdown.NewClient("ws://127.0.0.1:1")
	m := New(englishTestConfig(), Deps{Client: client})
	bv := decisionBattle(m, nil)
	bv.state().Request.Side.Pokemon[0].Condition = "0 fnt"
	bv.advanceSlot()
	if bv.slot != 1 {
		t.Fatal("fainted slot needs no input")
	}
	cmd := bv.chooseMove(1, m)
	if cmd == nil {
		t.Fatal("missing send")
	}
	m.Update(cmd())
	if bv.waiting || bv.sending || bv.lastError == "" || bv.slot != 1 {
		t.Fatal("failure left user waiting")
	}
	// A completion from an obsolete request must not overwrite a newer request.
	serial := bv.choiceSerial
	bv.clearWaiting()
	m.Update(choiceSentMsg{room: bv.room, serial: serial, note: "old"})
	if bv.waiting {
		t.Fatal("stale write completion applied")
	}
}

func TestCloseSendsLeaveAndReplayResetsWithoutFocusSteal(t *testing.T) {
	client, wire := challengeTestClient(t)
	m := New(englishTestConfig(), Deps{Client: client})
	room := "battle-gen9ou-123"
	init := showdown.RoomInit{Base: showdown.Base{RoomID: room}, Type: "battle"}
	m.handleShowdownEvent(init)
	m.battles[room].state().Log = []battle.LogEntry{{Text: "old"}}
	m.screen = screenLobby
	m.handleShowdownEvent(init)
	if len(m.battles[room].state().Log) != 0 || m.screen != screenLobby || len(m.order) != 1 {
		t.Fatal("rejoin duplicated history or stole focus")
	}
	cmd := m.dismissBattle(room)
	if cmd == nil {
		t.Fatal("missing leave")
	}
	m.Update(cmd())
	expectBattleCommand(t, wire, "|/leave "+room)
	m.handleShowdownEvent(init)
	if m.battles[room] != nil {
		t.Fatal("closed battle reopened")
	}
	cmd = m.submitSpectate("https://play.pokemonshowdown.com/" + room)
	m.Update(cmd())
	expectBattleCommand(t, wire, "|/join "+room)
	m.handleShowdownEvent(init)
	if m.battles[room] == nil {
		t.Fatal("explicit rejoin blocked")
	}
}

func TestBattleURLs(t *testing.T) {
	for _, input := range []string{"https://play.pokemonshowdown.com/battle-gen9ou-123", "https://play.pokemonshowdown.com/battle-gen9ou-123/?foo=1#bar", "replay.pokemonshowdown.com/gen9ou-123"} {
		if got := battleIDFromInput(input); got != "battle-gen9ou-123" {
			t.Fatalf("%q: %q", input, got)
		}
	}
	for _, input := range []string{"https://evil.test/replay.pokemonshowdown.com/gen9ou-123", "battle-gen9ou-123\n/forfeit", "https://play.pokemonshowdown.com/battle-gen9ou-123/extra"} {
		if got := battleIDFromInput(input); got != "" {
			t.Fatalf("accepted %q", input)
		}
	}
}

func TestLoginSuccessAndRemember(t *testing.T) {
	m := New(englishTestConfig(), Deps{})
	m.cfg.Remember = true
	m.handleShowdownEvent(showdown.AuthFailed{Err: errors.New("bad password")})
	for _, cmd := range m.commands() {
		if strings.HasPrefix(cmd.title, "Sign in to") {
			cmd.run(m)
		}
	}
	if m.prompt.value(2) != "y" {
		t.Fatal("remember preference lost")
	}
	m.handleShowdownEvent(showdown.UpdateUser{Name: "Player", Named: true})
	if m.authErr != "" {
		t.Fatal("stale login error")
	}
}

func TestFriendlyRematchTimerAndReplay(t *testing.T) {
	client, wire := challengeTestClient(t)
	m := New(englishTestConfig(), Deps{Client: client})
	bv := m.battleFor("battle-gen9randombattle-123")
	s := bv.state()
	s.Me, s.Tier, s.Format, s.Ended = "p1", "[Gen 9] Random Battle", "[Gen 9] Random Battle", true
	s.P2.Name = "Rival"
	cmd, _ := bv.handleKeyString("enter", m)
	if cmd == nil {
		t.Fatal("missing rematch")
	}
	m.Update(cmd())
	expectBattleCommand(t, wire, "|/utm null")
	expectBattleCommand(t, wire, "|/challenge Rival, gen9randombattle")
	for _, tc := range []struct{ key, command string }{{"ctrl+t", "/timer on"}, {"v", "/savereplay"}} {
		cmd, _ = bv.handleKeyString(tc.key, m)
		m.Update(cmd())
		expectBattleCommand(t, wire, bv.room+"|"+tc.command)
	}
	if strings.Contains(strings.Join(bv.renderResult(100), "\n"), "https://replay.") {
		t.Fatal("invented unsaved replay URL")
	}
	url := "https://replay.pokemonshowdown.com/gen9randombattle-123-secretpw"
	m.handleShowdownEvent(showdown.Popup{Message: `|html|<a href="` + url + `">` + url + `</a>`})
	if bv.replayURL != url {
		t.Fatalf("saved replay URL = %q", bv.replayURL)
	}
}

func TestTeamManagementRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := teams.Open(filepath.Join(dir, "teams.json"))
	if err != nil {
		t.Fatal(err)
	}
	client, wire := challengeTestClient(t)
	m := New(englishTestConfig(), Deps{Teams: store, Client: client})
	m.submitImport("=== [gen9ou] 원래 팀 ===\n\nPikachu\n- Thunderbolt")
	if len(store.Teams) != 1 || store.Teams[0].Format != "gen9ou" || store.Teams[0].Name != "원래 팀" {
		t.Fatal("import metadata lost")
	}
	m.teamAction("r")
	m.prompt.fields[0].value = "새 팀"
	m.submitPrompt()
	m.teamAction("e")
	path := filepath.Join(dir, "export.txt")
	m.prompt.fields[0].value = path
	m.submitPrompt()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := teams.ParseExport(string(raw))
	if err != nil || parsed.Name != "새 팀" || parsed.Format != "gen9ou" {
		t.Fatalf("export round trip: %+v %v", parsed, err)
	}
	cmd, _ := m.teamAction("enter")
	m.Update(cmd())
	expectBattleCommand(t, wire, "|/utm "+store.Teams[0].Pack())
	expectBattleCommand(t, wire, "|/search gen9ou")
	m.teamAction("d")
	if len(store.Teams) != 1 || m.confirm == "" {
		t.Fatal("deletion must require confirmation")
	}
	m.handleConfirmKey("y")
	reloaded, err := teams.Open(store.Path())
	if err != nil || len(reloaded.Teams) != 0 {
		t.Fatal("delete did not persist")
	}
}

func TestPreferencesAndQuitPaths(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := englishTestConfig()
	cfg.NoColor, cfg.Sidecar = true, true
	cfg.Keybindings = map[string]string{"z": "s"}
	m := New(cfg, Deps{})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if m.layout != LayoutSidecar {
		t.Fatal("sidecar ignored")
	}
	if out := m.View().Content; out != ansi.Strip(out) {
		t.Fatal("no_color ignored")
	}
	m.settingsCursor = 5
	m.handleSettingsKey("enter")
	loaded, _, err := config.Load()
	if err != nil || loaded.Notifications != m.cfg.Notifications || m.deps.Notifier.Mode != "osc" {
		t.Fatal("notification setting did not apply/save")
	}
	bv := decisionBattle(m, nil)
	m.active, m.screen = bv.room, screenBattle
	m.handleKey(keyMsg("z"))
	if bv.overlay != overlaySwitch {
		t.Fatal("key mapping ignored")
	}
	for _, cmd := range m.commands() {
		if cmd.title == "Quit" {
			if cmd.run(m) != nil || m.confirm == "" {
				t.Fatal("palette bypassed confirmation")
			}
		}
	}
	m.confirm, m.confirmAction = "", ""
	m.screen = screenLobby
	if cmd := m.handleLobbyKey("q"); cmd != nil || m.confirm == "" || m.formatFilter != "" {
		t.Fatal("lobby q did not ask to quit with active battle")
	}
	bv.curLayout, bv.curWidth = LayoutCinematic, 120
	m.cfg.Sprites.Size = "small"
	small, _ := bv.spriteCells()
	m.cfg.Sprites.Size = "large"
	large, _ := bv.spriteCells()
	if small >= large {
		t.Fatal("sprite size ignored")
	}
}

func TestKoreanTeamLayoutAndInputBoundaries(t *testing.T) {
	m := New(config.Default(), Deps{Teams: &teams.Store{Teams: []teams.Team{{Name: strings.Repeat("긴이름", 20), Format: "gen9ou", Pokemon: []teams.Pokemon{{Species: "Pikachu"}}}}}})
	for _, line := range strings.Split(m.renderTeams(46, 24), "\n") {
		if ansi.StringWidth(line) > 46 {
			t.Fatalf("team line exceeds terminal: %q", line)
		}
	}
	m.prompt = newPrompt("team-name", "", promptField{value: "피카츄"})
	m.handlePromptKey("backspace")
	if m.prompt.value(0) != "피카" {
		t.Fatal("backspace split Korean UTF-8")
	}
	m.prompt = promptState{}
	m.cfg.Keybindings = map[string]string{"z": "s"}
	bv := decisionBattle(m, nil)
	m.active, m.screen = bv.room, screenBattle
	bv.overlay = overlayChat
	m.handleKey(keyMsg("z"))
	if bv.chatInput != "z" {
		t.Fatal("key override rewrote chat text")
	}
}

func TestSelectedTeamAndNoUnrelatedFallback(t *testing.T) {
	m := New(englishTestConfig(), Deps{Teams: challengeTestTeams()})
	if got := m.teamForFormat(showdown.Format{ID: "gen9uu"}); got != "" {
		t.Fatal("unrelated format team selected")
	}
	m.deps.Teams.Teams = append(m.deps.Teams.Teams, teams.Team{Name: "Selected", Format: "gen9ou", Pokemon: []teams.Pokemon{{Species: "Gengar"}}})
	m.selectedTeam = 2
	if got := m.teamForFormat(showdown.Format{ID: "gen9ou"}); got != m.deps.Teams.Teams[2].Pack() {
		t.Fatal("explicit team selection ignored")
	}
}

func TestDoublesTargetPickerContinuesToSecondSlot(t *testing.T) {
	client, wire := challengeTestClient(t)
	m := New(englishTestConfig(), Deps{Client: client})
	bv := decisionBattle(m, nil)
	for _, side := range []*battle.Side{bv.state().P1, bv.state().P2} {
		side.Party = []*battle.Pokemon{{Name: "A", Active: true, Position: 0}, {Name: "B", Active: true, Position: 1}}
	}
	bv.state().Request.Active[0].Moves[0].Target = "normal"
	bv.state().Request.Active[1].Moves[0].Target = "normal"
	if cmd := bv.chooseMove(1, m); cmd != nil || bv.overlay != overlayTarget {
		t.Fatal("missing first target picker")
	}
	if cmd := bv.handleOverlayKey("2", m); cmd != nil || bv.slot != 1 || bv.overlay != overlayNone {
		t.Fatal("first target submitted incomplete turn")
	}
	bv.chooseMove(1, m)
	cmd := bv.handleOverlayKey("1", m)
	if cmd == nil {
		t.Fatal("second target did not complete turn")
	}
	m.Update(cmd())
	expectBattleCommand(t, wire, bv.room+"|/choose move 1 +2, move 1 +1|7")
}

func TestPreviewFailureAndNoCancel(t *testing.T) {
	m := New(englishTestConfig(), Deps{Client: showdown.NewClient("ws://127.0.0.1:1")})
	bv := decisionBattle(m, nil)
	bv.state().P1.Party = []*battle.Pokemon{{Name: "Pikachu"}, {Name: "Gengar"}}
	bv.state().TeamPreview = true
	bv.state().Request.TeamPreview = true
	cmd, _ := bv.handleKeyString("enter", m)
	if cmd == nil || !bv.sending {
		t.Fatal("preview not sent")
	}
	if duplicate, _ := bv.handleKeyString("enter", m); duplicate != nil {
		t.Fatal("duplicate preview sent")
	}
	m.Update(cmd())
	if bv.sending || bv.waiting || bv.lastError == "" {
		t.Fatal("failed preview left waiting")
	}
	bv.waiting = true
	bv.state().Request.NoCancel = true
	if cmd := bv.undo(); cmd != nil || !bv.waiting {
		t.Fatal("server noCancel ignored")
	}
}
