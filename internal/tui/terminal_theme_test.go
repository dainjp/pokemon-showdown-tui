package tui

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/unnipv/pokemon-slowdown/internal/config"
)

func TestTerminalThemeInheritsAllColors(t *testing.T) {
	theme := LoadTheme("terminal")
	check := func(name string, s lipgloss.Style) {
		t.Helper()
		if s.GetForeground() != (lipgloss.NoColor{}) || s.GetBackground() != (lipgloss.NoColor{}) || s.GetReverse() {
			t.Errorf("%s overrides terminal colors", name)
		}
	}
	value := reflect.ValueOf(theme)
	for i := 0; i < value.NumField(); i++ {
		if s, ok := value.Field(i).Interface().(lipgloss.Style); ok {
			check(value.Type().Field(i).Name, s)
		}
	}
	for k, s := range theme.Types {
		check("type "+k, s)
	}
	for k, s := range theme.TypesFg {
		check("move "+k, s)
	}
	for k, s := range theme.Statuses {
		check("status "+k, s)
	}
	if !theme.Selected.GetBold() || !theme.Selected.GetUnderline() {
		t.Fatal("selection loses its non-color cue")
	}
	if config.Default().Theme != "terminal" || LoadTheme("").Name != "terminal" {
		t.Fatal("terminal theme is not default")
	}
	if LoadTheme("midnight").Canvas.GetBackground() == (lipgloss.NoColor{}) {
		t.Fatal("explicit color themes no longer work")
	}
}

func TestTerminalThemeSwitchAppliesAndPersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := config.Default()
	cfg.Theme = "midnight"
	m := New(cfg, Deps{})
	bv := m.battleFor("battle-theme")
	found := false
	for _, cmd := range m.commands() {
		if cmd.title == "터미널 기본 색상 사용" {
			cmd.run(m)
			found = true
			break
		}
	}
	if !found || bv.theme.Name != "terminal" || m.theme.Name != "terminal" {
		t.Fatal("one-step terminal theme switch missing")
	}
	saved, _, err := config.Load()
	if err != nil || saved.Theme != "terminal" {
		t.Fatalf("saved preference: %s %v", saved.Theme, err)
	}
	if !strings.Contains(m.renderSettings(60, 24), "터미널 기본") {
		t.Fatal("untranslated theme label")
	}
}

func TestTerminalThemeRendersWithoutColorEscapes(t *testing.T) {
	m := New(config.Default(), Deps{})
	m.cfg.Sprites.Mode = "none"
	bv := m.battleFor("battle-theme")
	bv.state().Me = "p1"
	feedInfo(bv, "|switch|p1a: Donphan|Donphan|100/100\n|switch|p2a: Tsareena|Tsareena|100/100\n|-ability|p2a: Tsareena|Queenly Majesty")
	m.active = bv.room
	color := regexp.MustCompile(`\x1b\[[0-9;:]*m`)
	for _, size := range [][2]int{{46, 24}, {100, 40}} {
		m.width, m.height = size[0], size[1]
		m.layout = LayoutForSize(m.width, m.height)
		for _, screen := range []screen{screenLobby, screenBattle, screenSettings} {
			m.screen = screen
			out := m.render()
			if !utf8.ValidString(out) {
				t.Fatal("invalid Korean UTF-8")
			}
			for _, line := range strings.Split(out, "\n") {
				if ansi.StringWidth(line) > m.width {
					t.Fatalf("width overflow: %q", line)
				}
			}
			for _, sgr := range color.FindAllString(out, -1) {
				for _, code := range strings.Split(strings.TrimSuffix(strings.TrimPrefix(sgr, "\x1b["), "m"), ";") {
					if strings.Contains(code, ":") || code == "38" || code == "48" || code == "7" || (len(code) == 2 && ((code >= "30" && code <= "47") || (code >= "90" && code <= "97"))) {
						t.Fatalf("unexpected color/reverse sequence: %q", sgr)
					}
				}
			}
		}
	}
}
