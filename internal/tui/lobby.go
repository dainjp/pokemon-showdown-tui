package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"

	"github.com/unnipv/pokemon-slowdown/internal/showdown"
	"github.com/unnipv/pokemon-slowdown/internal/teams"
)

// handleKey routes a key press for the whole application.
func (m *Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.palette.open {
		return m, m.handlePaletteKey(key)
	}
	if m.prompt.open {
		return m, m.handlePromptKey(key)
	}
	if m.confirm != "" {
		return m, m.handleConfirmKey(key)
	}
	if m.helpOpen {
		if key == "esc" || key == "?" || key == "q" {
			m.helpOpen = false
		}
		return m, nil
	}

	textInput := m.screen == screenBattle && m.activeBattle() != nil && m.activeBattle().overlay == overlayChat
	if m.screen != screenSettings && !textInput {
		if target, ok := m.cfg.Keybindings[key]; ok {
			key = target
		}
	}
	switch key {
	case "f2":
		if m.screen == screenSettings {
			m.screen = m.settingsBack
		} else {
			m.settingsBack = m.screen
			m.screen = screenSettings
		}
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case ":":
		m.openPalette()
		return m, nil
	case "?":
		m.helpOpen = true
		return m, nil
	case "tab":
		if m.screen == screenSettings {
			return m, m.handleSettingsKey(key)
		}
		// Tab switches between open battles. It never starts or queues one.
		if m.screen == screenBattle {
			m.switchBattle()
		} else if len(m.order) > 0 {
			m.openBattle(m.order[len(m.order)-1])
		}
		return m, nil
	}

	switch m.screen {
	case screenBattle:
		if bv := m.activeBattle(); bv != nil {
			cmd, handled := bv.handleKeyString(key, m)
			if handled {
				return m, cmd
			}
		}
		if key == "q" {
			return m, m.requestQuit()
		}
	case screenTeams:
		return m, m.handleTeamsKey(key)
	case screenSettings:
		return m, m.handleSettingsKey(key)
	default:
		return m, m.handleLobbyKey(key)
	}
	return m, nil
}

func (m *Model) handleConfirmKey(key string) tea.Cmd {
	switch key {
	case "enter", "y":
		action := m.confirmAction
		m.confirm = ""
		m.confirmAction = ""
		switch action {
		case "delete-team":
			return m.deleteTeam()
		case "quit":
			return tea.Quit
		case "forfeit":
			if bv := m.activeBattle(); bv != nil && m.deps.Client != nil {
				room := bv.room
				client := m.deps.Client
				return func() tea.Msg {
					if err := client.SendTo(room, "/forfeit"); err != nil {
						return commandDoneMsg{err: err}
					}
					return commandDoneMsg{}
				}
			}
		}
	case "esc", "n", "q":
		m.confirm = ""
		m.confirmAction = ""
	}
	return nil
}

// ---------------------------------------------------------------------------
// Lobby
// ---------------------------------------------------------------------------

func (m *Model) handleLobbyKey(key string) tea.Cmd {
	items := m.filteredFormats()
	switch key {
	case "q":
		return m.requestQuit()
	case "up":
		m.formatCursor = clamp(m.formatCursor-1, 0, max(0, len(items)-1))
	case "down":
		m.formatCursor = clamp(m.formatCursor+1, 0, max(0, len(items)-1))
	case "pgup":
		m.formatCursor = clamp(m.formatCursor-10, 0, max(0, len(items)-1))
	case "pgdown":
		m.formatCursor = clamp(m.formatCursor+10, 0, max(0, len(items)-1))
	case "home":
		m.formatCursor = 0
	case "end":
		m.formatCursor = max(0, len(items)-1)
	case "enter":
		if m.formatCursor < len(items) {
			if m.challengeUser != "" {
				return m.challengeFormat(m.challengeUser, items[m.formatCursor])
			}
			return m.queueFormat(items[m.formatCursor])
		}
	case "tab":
		// Jump to the most recent battle, if there is one.
		if len(m.order) > 0 {
			m.openBattle(m.order[len(m.order)-1])
		}
	case "ctrl+t":
		m.screen = screenTeams
	case "backspace":
		if m.formatFilter != "" {
			runes := []rune(m.formatFilter)
			m.formatFilter = string(runes[:len(runes)-1])
			m.formatCursor = 0
		}
	case "esc":
		m.challengeUser = ""
		m.formatFilter = ""
		m.formatCursor = 0
		m.formatScroll = 0
	default:
		// Every printable key filters the list. Letter shortcuts would make
		// the search box unusable: typing "t" would open Teams instead of
		// searching for "Ting-Lu".
		if key == "space" {
			m.formatFilter += " "
			m.formatCursor = 0
			break
		}
		if r := printable(key); r != 0 {
			m.formatFilter += string(r)
			m.formatCursor = 0
		}
	}
	return nil
}

// formatLabel returns a format's display name, preferring the name the server
// sent over the curated fallback table.
func formatLabel(f showdown.Format) string {
	if f.Name != "" {
		return f.Name
	}
	return FormatName(f.ID)
}

// filteredFormats applies the fuzzy filter to the server's format list.
func (m *Model) filteredFormats() []showdown.Format {
	formats := m.formats
	if m.challengeUser != "" {
		formats = nil
		for _, f := range m.formats {
			if f.Challengeable {
				formats = append(formats, f)
			}
		}
	}
	if m.formatFilter == "" {
		return formats
	}
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = m.formatLabel(f) + " " + formatLabel(f) + " " + f.ID
	}
	matches := fuzzy.Find(m.formatFilter, names)
	out := make([]showdown.Format, 0, len(matches))
	for _, mt := range matches {
		out = append(out, formats[mt.Index])
	}
	return out
}

// queueFormat uploads the right team and starts a ladder search.
func (m *Model) queueFormat(f showdown.Format) tea.Cmd {
	if m.deps.Client == nil {
		return nil
	}
	client := m.deps.Client
	cmd := m.withTeam(f, func() error { return client.Search(f.ID) })
	if cmd != nil {
		m.pendingBattle = ""
		m.setToast(m.tr("Searching ") + m.formatLabel(f) + "…")
	}
	return cmd
}

// requeue starts a new search for a format, used from the end-of-battle screen
// so "play another" is one keystroke.
func (m *Model) requeue(formatID string) tea.Cmd {
	if formatID == "" {
		formatID = m.cfg.DefaultFormat
	}
	return m.queueFormat(m.formatInfo(formatID))
}

// teamForFormat picks a stored team to submit. Random formats need none.
func (m *Model) teamForFormat(f showdown.Format) string {
	if f.Random {
		return "null"
	}
	if m.deps.Teams == nil {
		return ""
	}
	all := m.deps.Teams.All()
	if len(all) == 0 {
		return ""
	}
	// Prefer the selected compatible team, then an exact format match.
	format, _, _ := strings.Cut(f.ID, "@@@")
	if m.selectedTeam >= 0 && m.selectedTeam < len(all) {
		t := all[m.selectedTeam]
		if (t.Format == format || t.Format == "") && len(t.Pokemon) > 0 {
			return t.Pack()
		}
	}
	for _, t := range all {
		if t.Format == format && len(t.Pokemon) > 0 {
			return t.Pack()
		}
	}
	for _, t := range all {
		if t.Format == "" && len(t.Pokemon) > 0 {
			return t.Pack()
		}
	}
	return ""
}

func (m *Model) renderLobby(width, height int) string {
	header := m.theme.Title.Render("Pokémon Slowdown") + "  " + m.theme.Muted.Render(m.tr("F2 settings · ? help"))
	header = truncate(header, width)
	height -= 2
	if m.layout == LayoutCompact {
		return header + "\n\n" + m.renderLobbyList(width, height)
	}

	listWidth := width * 3 / 5
	if listWidth < 30 {
		listWidth = 30
	}
	sideWidth := width - listWidth - 2

	left := m.renderLobbyList(listWidth, height)
	right := m.renderLobbySide(sideWidth, height)
	left = lipgloss.NewStyle().Width(listWidth).MaxWidth(listWidth).MaxHeight(height).Render(left)
	right = lipgloss.NewStyle().Width(sideWidth).MaxWidth(sideWidth).MaxHeight(height).Render(right)
	return header + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func (m *Model) renderLobbyList(width, height int) string {
	t := m.theme
	items := m.filteredFormats()

	var b strings.Builder
	title := m.tr("Find a battle")
	if m.challengeUser != "" {
		title = m.tr("Challenge format: ") + SanitizeLine(m.challengeUser)
	}
	b.WriteString(t.Title.Render(truncate(title, width)) + "\n")
	filter := m.formatFilter
	if filter == "" {
		filter = t.Dim.Render(m.tr("search formats…"))
	} else {
		filter = t.Fg.Render(filter) + t.Primary.Render("▌")
	}
	b.WriteString(t.Muted.Render("  / ") + filter + "\n\n")

	rows := height - 5
	if rows < 3 {
		rows = 3
	}
	if m.formatCursor < m.formatScroll {
		m.formatScroll = m.formatCursor
	}
	if m.formatCursor >= m.formatScroll+rows {
		m.formatScroll = m.formatCursor - rows + 1
	}
	end := m.formatScroll + rows
	if end > len(items) {
		end = len(items)
	}
	if len(items) == 0 {
		b.WriteString(t.Muted.Render(m.tr("  no formats matched\n")))
	}

	// Leave a fixed column for the tags, so a long name cannot run into them.
	nameW := width - 18
	if nameW < 12 {
		nameW = 12
	}
	if nameW > 40 {
		nameW = 40
	}

	for i := m.formatScroll; i < end; i++ {
		f := items[i]
		name := truncate(m.formatLabel(f), nameW)

		var tags []string
		if f.Random {
			tags = append(tags, t.Accent.Render(m.tr("random")))
		}
		if !f.Searchable {
			tags = append(tags, t.Dim.Render(m.tr("challenge")))
		}
		tagText := strings.Join(tags, " ")

		if i == m.formatCursor {
			row := padRight(name, nameW)
			if tagText != "" {
				row += "  " + tagText
			}
			b.WriteString(t.Selected.Width(width-2).Render("▸ "+row) + "\n")
			continue
		}
		row := padRight(t.Fg.Render(name), nameW)
		if tagText != "" {
			row += "  " + tagText
		}
		b.WriteString("  " + row + "\n")
	}
	return b.String()
}

func (m *Model) renderLobbySide(width, height int) string {
	t := m.theme
	if width < 10 {
		return ""
	}
	var b strings.Builder
	b.WriteString(t.Title.Render(m.tr("Status")) + "\n\n")
	if m.loggedIn {
		b.WriteString(t.Success.Render(SanitizeLine(m.username)) + "\n\n")
	}
	b.WriteString(t.Muted.Render(m.tr("Language")) + "  " + m.languageName() + "\n")
	b.WriteString(t.Muted.Render(m.tr("Sprites")) + "  " + m.onOff(m.cfg.Sprites.Mode != "none") + "\n")
	b.WriteString(t.Muted.Render(m.tr("Theme")) + "  " + m.cfg.Theme + "\n")
	b.WriteString(t.Dim.Render(m.tr("F2 to change settings")) + "\n\n")

	if len(m.searching) > 0 {
		b.WriteString(t.Accent.Render("● searching "+strings.Join(m.searching, ", ")) + "\n")
		b.WriteString(t.Dim.Render(m.tr("  : cancel search")) + "\n\n")
	}

	if len(m.games) > 0 {
		b.WriteString(t.Muted.Render(m.tr("Battles")) + "\n")
		rooms := make([]string, 0, len(m.games))
		for r := range m.games {
			rooms = append(rooms, r)
		}
		sort.Strings(rooms)
		for _, r := range rooms {
			title := SanitizeLine(m.games[r])
			line := "  " + t.Fg.Render(truncate(title, width-14))
			if bv := m.battles[r]; bv != nil && bv.state().Ended {
				// Tab skips finished battles; say so rather than leaving the
				// list looking like they are still playable.
				line += " " + t.Dim.Render("finished")
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	if len(m.challengesFrom) > 0 {
		b.WriteString(t.Warning.Render(m.tr("Challenges")) + "\n")
		for user, format := range m.challengesFrom {
			b.WriteString("  " + t.Fg.Render(SanitizeLine(user)) + " " +
				t.Muted.Render(FormatName(format)) + "\n")
		}
		b.WriteString(t.Dim.Render(m.tr("  : accept / reject")) + "\n\n")
	}
	for _, user := range sortedChallengeUsers(m.challengesTo) {
		b.WriteString(t.Muted.Render(m.tr("Challenging ")+SanitizeLine(user)+" · "+m.formatLabel(m.formatInfo(m.challengesTo[user]))) + "\n")
		b.WriteString(t.Dim.Render(m.tr("  : cancel challenge")) + "\n\n")
	}

	if m.conn != connOnline {
		b.WriteString(t.Warning.Render(m.tr("connection: ")+m.tr(m.conn.label())) + "\n")
	}
	if m.authErr != "" {
		b.WriteString(t.Danger.Render(truncate(m.authErr, width-2)) + "\n")
	}
	if !m.loggedIn {
		if m.cfg.Username != "" {
			b.WriteString(t.Warning.Render(m.tr("not signed in as ")+truncate(m.cfg.Username, width-20)) + "\n")
			b.WriteString(t.Dim.Render(m.tr("  : sign in")) + "\n")
		} else {
			b.WriteString(t.Dim.Render(m.tr("guest · : sign in for ladder ranking")) + "\n")
		}
	}

	b.WriteString("\n" + t.Muted.Render(m.tr("enter queue · ctrl+t teams · tab battle · : palette · ? help")))
	return b.String()
}

// ---------------------------------------------------------------------------
// Teams
// ---------------------------------------------------------------------------

func (m *Model) handleTeamsKey(key string) tea.Cmd {
	if cmd, handled := m.teamAction(key); handled {
		return cmd
	}
	switch key {
	case "esc":
		m.screen = screenLobby
	case "up", "k":
		m.teamCursor = clamp(m.teamCursor-1, 0, max(0, len(m.teamsList())-1))
	case "down", "j":
		m.teamCursor = clamp(m.teamCursor+1, 0, max(0, len(m.teamsList())-1))
	}
	return nil
}

func (m *Model) teamsList() []teams.Team {
	if m.deps.Teams == nil {
		return nil
	}
	return m.deps.Teams.All()
}

func (m *Model) renderTeams(width, height int) string {
	t := m.theme
	list := m.teamsList()
	var b strings.Builder
	b.WriteString(t.Title.Render(m.tr("Your teams")) + "\n\n")
	if len(list) == 0 {
		b.WriteString(t.Muted.Render(m.tr("  No teams stored yet.")) + "\n")
		b.WriteString(t.Dim.Render(m.tr("  Import one with : import team")) + "\n")
		return b.String()
	}
	rows := max(1, height-12)
	start := max(0, m.teamCursor-rows+1)
	for i := start; i < len(list) && i < start+rows; i++ {
		team := list[i]
		marker := "  "
		if i == m.teamCursor {
			marker = t.Accent.Render("▸ ")
		}
		b.WriteString(truncate(marker+t.Fg.Render(team.Name)+" "+t.Muted.Render(fmt.Sprintf("(%d) %s", len(team.Pokemon), team.Format)), width) + "\n")
	}
	if m.teamCursor < len(list) {
		b.WriteString("\n" + t.Muted.Render(m.tr("Preview")) + "\n")
		for _, p := range list[m.teamCursor].Pokemon {
			b.WriteString(truncate("  "+t.Fg.Render(m.name("pokemon", p.Display()))+" "+t.Muted.Render(m.name("item", p.Item)), width) + "\n")
		}
	}
	b.WriteString("\n" + t.Dim.Render(truncate(m.tr("Enter battle · R rename · F format · E export · D delete · Esc back"), width)))
	return b.String()
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	return ansi.Truncate(s, n, "…")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
