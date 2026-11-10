package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

type commandDoneMsg struct{ err error }
type choiceSentMsg struct {
	room   string
	serial int
	note   string
	undo   bool
	err    error
}

func (bv *battleView) passSlot(i int) bool {
	req := bv.state().Request
	if req == nil {
		return false
	}
	if req.Kind() == battle.RequestSwitch {
		if !req.MustSwitchAt(i) {
			return true
		}
		used := 0
		for _, c := range bv.draft {
			if c.Kind == "switch" {
				used++
			}
		}
		return used >= len(req.BenchSlots())
	}
	if req.Kind() != battle.RequestMove {
		return false
	}
	if i < len(req.Side.Pokemon) {
		p := req.Side.Pokemon[i]
		if strings.Contains(p.Condition, "fnt") || p.Commanding {
			return true
		}
	}
	ar := req.ActiveAt(i)
	return ar == nil || len(ar.Moves) == 0
}

func (bv *battleView) advanceSlot() {
	req := bv.state().Request
	if req == nil {
		return
	}
	for i := 0; i < req.SlotCount(); i++ {
		if _, ok := bv.draft[i]; !ok && !bv.passSlot(i) {
			bv.slot = i
			return
		}
	}
}

func (bv *battleView) undo() tea.Cmd {
	req := bv.state().Request
	if req == nil || req.Wait || bv.sending || (bv.waiting && (req.NoCancel || req.Side.NoCancel)) {
		return nil
	}
	if !bv.waiting {
		bv.draft = map[int]battle.ChoiceSlot{}
		bv.pendingMechanic = ""
		bv.overlay = overlayNone
		bv.advanceSlot()
		return nil
	}
	return bv.sendChoice(battle.UndoChoice(), "")
}

func (bv *battleView) roomCommand(command string) tea.Cmd {
	if bv.deps.Client == nil {
		return nil
	}
	room, client := bv.room, bv.deps.Client
	return func() tea.Msg { return commandDoneMsg{err: client.SendTo(room, command)} }
}

func (bv *battleView) formatID() string {
	id := strings.TrimPrefix(bv.room, "battle-")
	if format, _, ok := strings.Cut(id, "-"); ok {
		return format
	}
	if bv.state().Tier != "" {
		return showdown.ToID(bv.state().Tier)
	}
	return bv.owner.cfg.DefaultFormat
}
