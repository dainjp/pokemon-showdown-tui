package tui

import (
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
	"github.com/unnipv/pokemon-slowdown/internal/sprites"
)

func TestKoreanCatalogueAndNicknameBoundaries(t *testing.T) {
	m := New(config.Default(), Deps{})
	for _, tc := range []struct{ category, english, korean string }{
		{"pokemon", "Garchomp", "한카리아스"},
		{"pokemon", "Nidoran-F", "니드런♀"},
		{"pokemon", "Nidoran-M", "니드런♂"},
		{"move", "Thunderbolt", "10만볼트"},
		{"ability", "Intimidate", "위협"},
		{"item", "Focus Sash", "기합의띠"},
		{"type", "electric", "전기"},
		{"pokemon", "Unreleased Custommon", "Unreleased Custommon"},
	} {
		if got := m.name(tc.category, tc.english); got != tc.korean {
			t.Errorf("%s %q = %q, want %q", tc.category, tc.english, got, tc.korean)
		}
	}
	bv := m.battleFor("battle-local-test")
	if got := bv.pokemonName(&battle.Pokemon{Name: "Sparky", Species: "Pikachu"}); got != "Sparky" {
		t.Fatalf("nickname changed: %q", got)
	}
	if got := bv.pokemonName(&battle.Pokemon{Name: "Pikachu", Species: "Pikachu"}); got != "피카츄" {
		t.Fatalf("species not translated: %q", got)
	}
	m.cfg.Language = "en"
	if got := m.name("pokemon", "Garchomp"); got != "Garchomp" {
		t.Fatalf("English fallback = %q", got)
	}
}

func TestKoreanBattleLogKeepsPlayerNamesAndProtocolUntouched(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-local-test")
	bv.apply(showdown.BattleMove{User: "p1a: Pikachu", Move: "Thunderbolt"})
	entry := bv.state().Log[len(bv.state().Log)-1]
	if entry.Text != "Pikachu used Thunderbolt!" {
		t.Fatalf("underlying English log changed: %q", entry.Text)
	}
	if got := bv.logText(entry); got != "피카츄의 10만볼트!" {
		t.Fatalf("localized log = %q", got)
	}
	bv.apply(showdown.BattleWin{User: "Garchomp"})
	entry = bv.state().Log[len(bv.state().Log)-1]
	if got := bv.logText(entry); got != "Garchomp 승리!" {
		t.Fatalf("player name translated: %q", got)
	}
	for _, entry := range bv.state().Log {
		if strings.Contains(bv.logText(entry), "%!") {
			t.Fatalf("bad localized format: %q", bv.logText(entry))
		}
	}
}

func TestRememberIsAToggleAndPasswordStaysOutOfSettings(t *testing.T) {
	keyring := newKeyringStub()
	m, dir := signInModel(t, keyring)
	m.prompt = newPrompt("login", "Sign in",
		promptField{label: "Username", value: "local-test"},
		promptField{label: "Password", value: "test-only-password", secret: true},
		promptField{label: "Remember", value: "n", toggle: true},
	)
	m.prompt.index = 2
	m.handlePromptKey("y")
	if m.prompt.value(2) != "y" {
		t.Fatal("Y appended to N instead of selecting yes")
	}
	m.handlePromptKey("tab")
	if m.prompt.index != 0 {
		t.Fatal("Tab did not move to the first field")
	}
	m.handlePromptKey("shift+tab")
	if m.prompt.index != 2 {
		t.Fatal("Shift+Tab did not move back")
	}
	m.handlePromptKey("enter")
	if keyring.stored["local-test"] != "test-only-password" {
		t.Fatal("remember did not use the keychain")
	}
	_ = dir // signInModel scopes the config directory to the test.
	loaded, _, err := config.Load()
	if err != nil || !loaded.Remember {
		t.Fatalf("remember was not persisted: %v", err)
	}
}

func TestSpriteAndThemeSettingsApplyToAnOpenBattleAndPersist(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := config.Default()
	cfg.Sprites.Mode = "blocks"
	renderer, err := sprites.NewRenderer(sprites.ModeBlocks, nil)
	if err != nil {
		t.Fatal(err)
	}
	m := New(cfg, Deps{Renderer: renderer})
	defer m.Cleanup(io.Discard)
	bv := m.battleFor("battle-local-test")
	m.toggleSprites()
	if m.cfg.Sprites.Mode != "none" || bv.deps.Renderer != m.deps.Renderer {
		t.Fatal("sprite setting did not reach the open battle")
	}
	m.toggleSprites()
	if m.cfg.Sprites.Mode != "blocks" {
		t.Fatal("sprite toggle lost the selected backend")
	}
	m.cycleTheme()
	if bv.theme.Name != m.cfg.Theme {
		t.Fatal("theme change did not reach the open battle")
	}
	loaded, _, err := config.Load()
	if err != nil || loaded.Theme != m.cfg.Theme || loaded.Sprites.Mode != "blocks" {
		t.Fatalf("preferences did not survive a reload: %v", err)
	}
}

func TestKoreanViewsFitTerminalCells(t *testing.T) {
	m := New(config.Default(), Deps{})
	feedFixtureUntilTurn(t, m, "gen9-singles.txt", 4)
	m.conn = connOnline
	for _, size := range [][2]int{{32, 12}, {40, 16}, {60, 24}, {80, 30}, {120, 40}} {
		m.width, m.height = size[0], size[1]
		m.layout = LayoutForSize(m.width, m.height)
		for _, screen := range []screen{screenLobby, screenSettings, screenBattle} {
			m.screen = screen
			out := m.render()
			if !utf8.ValidString(out) {
				t.Fatalf("invalid Korean UTF-8 at %dx%d", m.width, m.height)
			}
			if rows := lipgloss.Height(out); rows > m.height {
				t.Errorf("screen %d is %d rows at %dx%d", screen, rows, m.width, m.height)
			}
			for _, line := range strings.Split(out, "\n") {
				if width := lipgloss.Width(line); width > m.width {
					t.Errorf("screen %d line is %d cells at %dx%d: %q", screen, width, m.width, m.height, line)
				}
			}
		}
	}
	for _, width := range []int{1, 3, 6, 9} {
		text := truncate("한카리아스와 피카츄", width)
		if !utf8.ValidString(text) || lipgloss.Width(text) > width {
			t.Errorf("bad CJK truncation at %d: %q", width, text)
		}
	}
}

func TestNarrowBattleFooterPreservesActionsAndConnection(t *testing.T) {
	m := New(config.Default(), Deps{})
	m.width = 60
	m.screen = screenBattle
	m.conn = connOnline
	m.username = "AnUnusuallyLongPlayerName"
	out := stripANSI(m.renderStatus())
	for _, want := range []string{"기술", "교체", "F2", ":", "연결됨"} {
		if !strings.Contains(out, want) {
			t.Errorf("narrow footer hid %q: %q", want, out)
		}
	}
	if strings.Contains(out, "…") || lipgloss.Width(out) > m.width {
		t.Fatalf("narrow footer clipped: %q", out)
	}
	m.width = 32
	out = stripANSI(m.renderStatus())
	if !strings.Contains(out, "교체") || !strings.Contains(out, ":") || !strings.Contains(out, "연결됨") || lipgloss.Width(out) > 32 {
		t.Fatalf("compact footer lost navigation: %q", out)
	}
}

func TestKoreanBackspaceKeepsWholeCharacters(t *testing.T) {
	m := New(config.Default(), Deps{})
	m.palette.filter = "설정"
	m.handlePaletteKey("backspace")
	if m.palette.filter != "설" {
		t.Fatalf("menu search corrupted Korean text: %q", m.palette.filter)
	}
	bv := m.battleFor("battle-local-test")
	bv.overlay = overlayChat
	bv.chatInput = "안녕"
	bv.handleOverlayKey("backspace", m)
	if bv.chatInput != "안" {
		t.Fatalf("chat editor corrupted Korean text: %q", bv.chatInput)
	}
}
