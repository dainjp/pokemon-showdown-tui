package tui

import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/config"
	"regexp"
	"strings"
	"testing"
)

func TestHealingAbilityOwner(t *testing.T) {
	for _, ability := range []string{"Water Absorb", "Volt Absorb", "Dry Skin", "Earth Eater", "Hospitality"} {
		t.Run(ability, func(t *testing.T) {
			m := New(config.Default(), Deps{})
			bv := m.battleFor("battle-review")
			bv.state().Me = "p1"
			sourceID := "p1a: Source"
			if ability == "Hospitality" {
				sourceID = "p2b: Source"
			}
			feedInfo(bv, "|switch|"+sourceID+"|Sinistcha|100/100\n|switch|p2a: Receiver|Vaporeon|50/100")
			source, receiver := bv.state().Find(sourceID), bv.state().Find("p2a: Receiver")
			source.Ability = "Unnerve"
			feedInfo(bv, "|-heal|p2a: Receiver|75/100|[from] ability: "+ability+"|[of] "+sourceID)
			t.Logf("source ability=%q; receiver ability=%q", source.Ability, receiver.Ability)
			if ability == "Hospitality" {
				if source.Ability != ability || receiver.Ability != "" {
					t.Error("Hospitality should belong to source")
				}
			} else if receiver.Ability != ability || source.Ability != "Unnerve" {
				t.Error("absorption ability assigned to attacker instead of receiver")
			}
		})
	}
}

func TestShortBattlePreservesMovesAndHeight(t *testing.T) {
	moveLine := regexp.MustCompile(`^\s+[1-4] `)
	count := 0
	for _, width := range []int{32, 46, 60, 80, 100} {
		for _, height := range []int{12, 14, 18, 24} {
			for _, slots := range []int{1, 2} {
				for _, waiting := range []bool{false, true} {
					t.Run(fmt.Sprintf("%dx%d_slots%d_wait%t", width, height, slots, waiting), func(t *testing.T) {
						cfg := config.Default()
						cfg.Sprites.Mode = "none"
						m := New(cfg, Deps{})
						bv := decisionBattle(m, nil)
						m.active, m.screen = bv.room, screenBattle
						m.width, m.height = width, height
						m.layout = LayoutForSize(width, height)
						m.toast = "전송 상태 안내"
						m.connectionNotice = "연결 상태 안내"
						bv.waiting = waiting
						bv.waitNote = strings.Repeat("Earthquake + Ice Shard ", 8)
						bv.lastError = "이전 선택 오류"
						bv.state().Request.Active = bv.state().Request.Active[:slots]
						moves := []battle.MoveRequest{{Move: "Tackle", PP: 35, MaxPP: 35}, {Move: "Protect", PP: 10, MaxPP: 10}, {Move: "Ice Shard", PP: 30, MaxPP: 30}, {Move: "Earthquake", PP: 10, MaxPP: 10}}
						for i := range bv.state().Request.Active {
							bv.state().Request.Active[i].Moves = moves
						}
						for _, side := range []*battle.Side{bv.state().P1, bv.state().P2} {
							for i := 0; i < slots; i++ {
								side.Party = append(side.Party, &battle.Pokemon{Name: "Pikachu", Species: "Pikachu", Active: true, Position: i, SideID: side.ID, HP: 100, MaxHP: 100, HPPercent: 100, Ability: "Static", Item: "Light Ball"})
							}
						}
						out := ansi.Strip(m.render())
						lines := strings.Split(out, "\n")
						visible := 0
						for i, line := range lines {
							if i < height && moveLine.MatchString(line) {
								visible++
							}
						}
						if visible != 4 || len(lines) > height {
							t.Fatalf("moves=%d, rows=%d/%d:\n%s", visible, len(lines), height, out)
						}
						count++
					})
				}
			}
		}
	}
	t.Logf("verified %d rendering cases", count)
}
