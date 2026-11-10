package tui

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

func sortedChallengeUsers(challenges map[string]string) []string {
	users := make([]string, 0, len(challenges))
	for user := range challenges {
		users = append(users, user)
	}
	sort.Strings(users)
	return users
}

func (m *Model) formatInfo(id string) showdown.Format {
	base, _, _ := strings.Cut(id, "@@@")
	for _, f := range m.formats {
		if f.ID == showdown.ToID(base) {
			f.ID = id
			return f
		}
	}
	return showdown.Format{
		ID: id, Random: strings.Contains(base, "random"),
		Searchable: true, Challengeable: true,
	}
}

func (m *Model) withTeam(f showdown.Format, send func() error) tea.Cmd {
	if m.deps.Client == nil {
		return nil
	}
	packed := m.teamForFormat(f)
	if packed == "" {
		m.setToast(m.tr("Import a team first: : menu → Import a team."))
		return nil
	}
	return m.battleCommand(packed, send)
}

// Only one team/request pair may be in flight, so rapid actions cannot replace
// the uploaded team before its search, challenge or acceptance command.
func (m *Model) battleCommand(packed string, send func() error) tea.Cmd {
	if m.deps.Client == nil || m.battleRequestPending {
		return nil
	}
	m.battleRequestPending = true
	client := m.deps.Client
	return func() tea.Msg {
		if packed != "" {
			if err := client.SendTeam(packed); err != nil {
				return battleRequestDoneMsg{err: err}
			}
		}
		return battleRequestDoneMsg{err: send()}
	}
}

func (m *Model) challengeFormat(user string, f showdown.Format) tea.Cmd {
	client := m.deps.Client
	cmd := m.withTeam(f, func() error { return client.Challenge(user, f.ID) })
	if cmd != nil {
		m.challengeUser = ""
		m.setToast(m.tr("Challenging ") + user + " · " + m.formatLabel(f) + "…")
	}
	return cmd
}

func (m *Model) acceptChallenge(user, format string) tea.Cmd {
	client := m.deps.Client
	cmd := m.withTeam(m.formatInfo(format), func() error { return client.AcceptChallenge(user) })
	if cmd != nil {
		m.setToast(m.tr("Accepting challenge from ") + user + "…")
	}
	return cmd
}

func (m *Model) updateChallenge(e showdown.ChallengeUpdated) {
	self := showdown.ToID(showdown.ParseIdentity(m.username).Name)
	if self == "" || e.Sender == "" || e.Receiver == "" {
		return
	}
	other, outgoing := e.Sender, false
	if showdown.ToID(e.Sender) == self {
		other, outgoing = e.Receiver, true
	} else if showdown.ToID(e.Receiver) != self {
		return
	}
	for _, challenges := range []map[string]string{m.challengesFrom, m.challengesTo} {
		for user := range challenges {
			if showdown.ToID(user) == showdown.ToID(other) {
				delete(challenges, user)
			}
		}
	}
	if e.Format == "" {
		return
	}
	format := e.TeamFormat
	if format == "" {
		format = e.Format
	}
	if outgoing {
		m.challengesTo[other] = format
	} else {
		m.challengesFrom[other] = format
		m.setToast(m.tr("Challenge from ") + other + " · " + m.formatLabel(m.formatInfo(format)))
	}
}
