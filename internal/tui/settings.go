package tui

import (
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/unnipv/pokemon-slowdown/internal/sprites"
)

func (m *Model) languageName() string {
	if m.cfg.Language == "ko" {
		return "한국어"
	}
	return "English"
}

func (m *Model) onOff(enabled bool) string {
	if enabled {
		return m.tr("On")
	}
	return m.tr("Off")
}

func (m *Model) savePreferences() {
	if err := m.cfg.Save(""); err != nil {
		m.setToast(m.tr("Could not save settings: ") + err.Error())
		return
	}
	m.setToast(m.tr("Settings saved."))
}

func (m *Model) toggleSprites() tea.Cmd {
	previous := m.cfg.Sprites.Mode
	if previous == "none" {
		m.cfg.Sprites.Mode = m.savedSpriteMode
	} else {
		m.savedSpriteMode = previous
		m.cfg.Sprites.Mode = "none"
	}
	if err := m.rebuildRenderer(); err != nil {
		m.cfg.Sprites.Mode = previous
		m.setToast(m.tr("Could not change sprites: ") + err.Error())
		return nil
	}
	m.savePreferences()
	if bv := m.activeBattle(); bv != nil {
		return tea.Batch(m.loadSprites(bv)...)
	}
	return nil
}

func (m *Model) cycleTheme() {
	m.cfg.Theme = nextTheme(m.cfg.Theme)
	m.theme = LoadTheme(m.cfg.Theme)
	for _, bv := range m.battles {
		bv.theme = m.theme
	}
	m.savePreferences()
}

func (m *Model) handleSettingsKey(key string) tea.Cmd {
	switch key {
	case "esc", "q":
		m.screen = m.settingsBack
	case "up", "shift+tab":
		m.settingsCursor = (m.settingsCursor + 3) % 4
	case "down", "tab":
		m.settingsCursor = (m.settingsCursor + 1) % 4
	case "enter", "space", "left", "right":
		switch m.settingsCursor {
		case 0:
			if m.cfg.Language == "ko" {
				m.cfg.Language = "en"
			} else {
				m.cfg.Language = "ko"
			}
		case 1:
			return m.toggleSprites()
		case 2:
			m.cfg.Sprites.Animate = !m.cfg.Sprites.Animate
		case 3:
			m.cycleTheme()
			return nil
		}
		m.savePreferences()
	}
	return nil
}

func (m *Model) renderSettings(width, height int) string {
	t := m.theme
	labels := []string{m.tr("Language"), m.tr("Sprites"), m.tr("Animation"), m.tr("Theme")}
	values := []string{m.languageName(), m.onOff(m.cfg.Sprites.Mode != "none"), m.onOff(m.cfg.Sprites.Animate), m.cfg.Theme}
	lines := []string{t.Title.Render(m.tr("Settings")), t.Muted.Render(m.tr("Changes are saved automatically.")), ""}
	for i, label := range labels {
		row := fmt.Sprintf("  %s  %s", padRight(label, 12), values[i])
		row = truncate(row, width-2)
		if i == m.settingsCursor {
			row = t.Selected.Width(width).Render("▸" + strings.TrimPrefix(row, " "))
		} else {
			row = t.Fg.Render(row)
		}
		lines = append(lines, row)
	}
	lines = append(lines, "", t.Muted.Render(truncate(m.tr("↑/↓ select · Enter change · Esc back"), width)))
	if height > len(lines) {
		lines = append(lines, make([]string, height-len(lines))...)
	}
	return strings.Join(lines, "\n")
}

// Cleanup releases whichever renderer is active after runtime settings changes.
func (m *Model) Cleanup(out io.Writer) {
	if m.deps.Renderer != nil {
		sprites.Cleanup(out, m.deps.Renderer.Capabilities())
		_ = m.deps.Renderer.Close()
	}
}
