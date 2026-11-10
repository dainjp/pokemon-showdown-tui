package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/dex"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
	"github.com/unnipv/pokemon-slowdown/internal/sprites"
	"github.com/unnipv/pokemon-slowdown/internal/storage"
)

// Opt-in visual evidence uses a recorded battle, never a live opponent or account.
func TestPersonalPreview(t *testing.T) {
	dir := os.Getenv("SLOWDOWN_PREVIEW_DIR")
	if dir == "" {
		t.Skip("set SLOWDOWN_PREVIEW_DIR to render the recorded-battle previews")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	d := dex.New(storage.DexDir())
	if err := d.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	renderer, err := sprites.NewRenderer(sprites.ModeBlocks, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Sprites.Mode = "blocks"
	cfg.Sprites.Animate = false
	m := New(cfg, Deps{Dex: d, Renderer: renderer})
	m.width, m.height = 120, 38
	m.layout = LayoutForSize(m.width, m.height)
	feedFixtureUntilTurn(t, m, "gen9-singles.txt", 4)
	m.username, m.loggedIn, m.conn = "Trainer", true, connOnline
	m.formats = []showdown.Format{
		{ID: "gen9randombattle", Name: "[Gen 9] Random Battle", Random: true, Searchable: true},
		{ID: "gen9randomdoublesbattle", Name: "[Gen 9] Random Doubles Battle", Random: true, Searchable: true},
		{ID: "gen9ou", Name: "[Gen 9] OU", Searchable: true},
		{ID: "gen9doublesou", Name: "[Gen 9] Doubles OU", Searchable: true},
		{ID: "gen9bssregi", Name: "[Gen 9] BSS Reg I", Searchable: false},
		{ID: "gen9customgame", Name: "[Gen 9] Custom Game", Searchable: false},
	}
	write := func(name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name+".ansi"), []byte(m.render()), 0644); err != nil {
			t.Fatal(err)
		}
	}
	m.screen = screenLobby
	write("lobby")
	m.screen = screenSettings
	m.settingsCursor = 1
	write("settings")
	m.prompt = newPrompt("login", m.tr("Sign in to Pokémon Showdown"),
		promptField{label: m.tr("Username")}, promptField{label: m.tr("Password"), secret: true},
		promptField{label: m.tr("Remember"), toggle: true, value: "y"})
	m.prompt.index = 2
	m.screen = screenLobby
	write("login")
	m.prompt = promptState{}
	bv := m.activeBattle()
	manager := sprites.NewManager(storage.SpriteDir())
	m.deps.Sprites = manager
	bv.deps.Sprites = manager
	for _, command := range m.loadSprites(bv) {
		if command != nil {
			m.Update(command())
		}
	}
	m.screen = screenBattle
	if !strings.ContainsAny(m.render(), "▀▄") {
		t.Fatal("preview did not render the loaded sprites")
	}
	write("battle-sprites-on")
	m.toggleSprites()
	if strings.ContainsAny(m.render(), "▀▄") || strings.Contains(m.render(), "····") {
		t.Fatal("sprites off left an image or placeholder")
	}
	write("battle-sprites-off")
	m.width, m.height = 60, 24
	m.layout = LayoutForSize(m.width, m.height)
	write("battle-narrow")
	m.width, m.height = 120, 38
	m.layout = LayoutForSize(m.width, m.height)
	m.screen = screenLobby
	m.handleShowdownEvent(showdown.AccessRestricted{})
	write("network-restriction")
}
