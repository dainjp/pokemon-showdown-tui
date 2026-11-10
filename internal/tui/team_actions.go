package tui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

func (m *Model) requestQuit() tea.Cmd {
	if m.screen == screenBattle || len(m.liveBattles()) > 0 {
		m.confirm, m.confirmAction = m.tr("Quit pokemon slowdown?"), "quit"
		return nil
	}
	return tea.Quit
}

func (m *Model) teamAction(key string) (tea.Cmd, bool) {
	if m.deps.Teams == nil {
		return nil, false
	}
	team, ok := m.deps.Teams.Get(m.teamCursor)
	if !ok {
		return nil, false
	}
	switch key {
	case "r":
		m.prompt = newPrompt("team-name", m.tr("Rename team"), promptField{label: m.tr("Name"), value: team.Name})
	case "f":
		m.prompt = newPrompt("team-format", m.tr("Set team format"), promptField{label: m.tr("Format"), value: team.Format})
	case "e":
		m.prompt = newPrompt("team-export", m.tr("Export team"), promptField{label: m.tr("New file path")})
	case "d", "delete":
		m.confirm, m.confirmAction = fmt.Sprintf(m.tr("Delete team %s?"), team.Name), "delete-team"
	case "enter":
		m.selectedTeam = m.teamCursor
		if team.Format == "" {
			m.prompt = newPrompt("team-format", m.tr("Set team format"), promptField{label: m.tr("Format")})
			m.setToast(m.tr("Set a format, then press Enter to battle."))
			return nil, true
		}
		return m.queueFormat(m.formatInfo(team.Format)), true
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) deleteTeam() tea.Cmd {
	if m.deps.Teams == nil {
		return nil
	}
	if err := m.deps.Teams.Delete(m.teamCursor); err != nil {
		m.setToast(err.Error())
		return nil
	}
	if m.selectedTeam == m.teamCursor {
		m.selectedTeam = -1
	} else if m.selectedTeam > m.teamCursor {
		m.selectedTeam--
	}
	m.teamCursor = clamp(m.teamCursor, 0, max(0, len(m.teamsList())-1))
	m.setToast(m.tr("Team deleted."))
	return nil
}

func (m *Model) submitTeamPrompt(p promptState) tea.Cmd {
	value := strings.TrimSpace(p.value(0))
	if value == "" {
		m.setToast(m.tr("Enter a value."))
		return nil
	}
	if p.kind == "default-format" {
		if showdown.ToID(value) != value {
			m.setToast(m.tr("Enter a format ID, e.g. gen9ou."))
			return nil
		}
		m.cfg.DefaultFormat = value
		m.savePreferences()
		return nil
	}
	if m.deps.Teams == nil {
		return nil
	}
	team, ok := m.deps.Teams.Get(m.teamCursor)
	if !ok {
		return nil
	}
	switch p.kind {
	case "team-name":
		team.Name = value
	case "team-format":
		if showdown.ToID(value) != value {
			m.setToast(m.tr("Enter a format ID, e.g. gen9ou."))
			return nil
		}
		team.Format = value
	case "team-export":
		// Refuse to overwrite an existing file; exporting should never destroy data.
		f, err := os.OpenFile(value, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err == nil {
			header := team.Name
			if team.Format != "" {
				header = "[" + team.Format + "] " + header
			}
			_, err = fmt.Fprint(f, "=== "+header+" ===\n\n"+team.Export())
			if closeErr := f.Close(); err == nil {
				err = closeErr
			}
		}
		if err != nil {
			m.setToast(m.tr("Could not export team: ") + err.Error())
		} else {
			m.setToast(m.tr("Exported team to ") + value)
		}
		return nil
	}
	if err := m.deps.Teams.Put(m.teamCursor, team); err != nil {
		m.setToast(err.Error())
	} else {
		m.setToast(m.tr("Team saved."))
	}
	return nil
}
