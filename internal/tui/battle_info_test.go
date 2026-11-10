package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

func feedInfo(bv *battleView, raw string) {
	for _, f := range showdown.SplitFrames(raw) {
		for _, ev := range showdown.Parse(f) {
			bv.apply(ev)
		}
	}
}

func TestRevealedEquipmentAndEffectLogs(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-info")
	bv.state().Me = "p1"
	feedInfo(bv, "|switch|p1a: Donphan|Donphan|100/100\n|switch|p2a: Tsareena|Tsareena|100/100")
	foe := bv.state().Find("p2a: Tsareena")
	if !strings.Contains(bv.itemLabel(foe), "?") {
		t.Fatal("unrevealed item treated as absent")
	}
	feedInfo(bv, "|-activate|p2a: Tsareena|ability: Queenly Majesty")
	got := bv.logText(bv.state().Log[len(bv.state().Log)-1])
	if got != "상대 달코퀸의 특성: 여왕의위엄 발동!" || foe.Ability != "Queenly Majesty" {
		t.Fatalf("activation: %s", got)
	}
	feedInfo(bv, "|-heal|p2a: Tsareena|100/100|[from] item: Leftovers")
	got = bv.logText(bv.state().Log[len(bv.state().Log)-1])
	if foe.Item != "Leftovers" || !strings.Contains(got, "도구: 먹다남은음식") {
		t.Fatalf("heal item: %s", got)
	}
	feedInfo(bv, "|-enditem|p2a: Tsareena|Leftovers")
	if got := bv.itemLabel(foe); !strings.Contains(got, "없음") || !strings.Contains(got, "먹다남은음식") || strings.Contains(got, "?") {
		t.Fatalf("consumed item: %s", got)
	}
	feedInfo(bv, "|-endability|p2a: Tsareena")
	if got := bv.abilityLabel(foe); !strings.Contains(got, "여왕의위엄") || !strings.Contains(got, "무효화") {
		t.Fatalf("suppressed: %s", got)
	}
	feedInfo(bv, "|-ability|p2a: Tsareena|Sweet Veil")
	if foe.AbilitySuppressed || foe.Ability != "Sweet Veil" {
		t.Fatal("ability replacement stale")
	}
	feedInfo(bv, "|-damage|p1a: Donphan|90/100|[from] Stealth Rock")
	if got := bv.logText(bv.state().Log[len(bv.state().Log)-1]); !strings.Contains(got, "[스텔스록]") {
		t.Fatalf("unprefixed effect: %s", got)
	}
	// Damage on us reveals the opponent's ability, not our own.
	feedInfo(bv, "|-damage|p1a: Donphan|90/100|[from] ability: Rough Skin|[of] p2a: Tsareena")
	if foe.Ability != "Rough Skin" || bv.state().Find("p1a: Donphan").Ability == "Rough Skin" {
		t.Fatal("damage source attributed to victim")
	}
}

func TestOwnRequestAndUnknownOpponentEquipment(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-info")
	feedInfo(bv, `|request|{"wait":true,"side":{"id":"p1","pokemon":[{"ident":"p1: Donphan","details":"Donphan","condition":"100/100","active":true,"baseAbility":"sturdy","item":""}]}}`)
	mine := bv.state().P1.Party[0]
	if bv.itemLabel(mine) != "없음" || bv.abilityLabel(mine) != "옹골참" {
		t.Fatalf("own request: %s %s", bv.itemLabel(mine), bv.abilityLabel(mine))
	}
	foe := &battle.Pokemon{SideID: "p2", Species: "Tsareena", BaseAbility: "Queenly Majesty"}
	if !strings.Contains(bv.abilityLabel(foe), "?") {
		t.Fatal("opponent hidden ability revealed from guess")
	}
}

func TestBattleDetailsWidthsAndInspectScrolling(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-info")
	bv.state().Me = "p1"
	feedInfo(bv, "|switch|p1a: Donphan|Donphan|100/100\n|switch|p2a: Tsareena|Tsareena|100/100\n|-ability|p1a: Donphan|Sturdy\n|-item|p1a: Donphan|Assault Vest\n|-ability|p2a: Tsareena|Queenly Majesty\n|-item|p2a: Tsareena|Leftovers")
	for _, width := range []int{32, 46, 70, 100} {
		out := strings.Join(bv.renderField(width, LayoutFor(width)), "\n")
		for _, line := range strings.Split(out, "\n") {
			if !utf8.ValidString(line) || ansi.StringWidth(line) > width {
				t.Fatalf("width %d: %q", width, line)
			}
		}
		if !strings.Contains(out, "특성:") || !strings.Contains(out, "도구:") {
			t.Fatalf("metadata missing at %d", width)
		}
	}
	p := bv.state().P1.Party[0]
	p.Moves = []battle.Move{{Name: "Ice Shard", ID: "iceshard"}, {Name: "Earthquake", ID: "earthquake"}}
	bv.curHeight = 18
	bv.inspectPokemon(p, overlayNone)
	top := ansi.Strip(bv.renderInspectOverlay(46))
	bv.handleOverlayKey("end", m)
	bottom := ansi.Strip(bv.renderInspectOverlay(46))
	if bv.inspectScroll == 0 || top == bottom || !strings.Contains(bottom, "얼음뭉치") {
		t.Fatalf("cannot reach moves in short terminal:\n%s", bottom)
	}
	bv.handleOverlayKey("up", m)
	if bv.inspectScroll != 0 {
		t.Fatal("scroll not reset on Pokémon change")
	}
}

func TestBoostStatsUseFloorAndExplainLimits(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-info")
	p := &battle.Pokemon{Stats: map[string]int{"spe": 101}, Boosts: map[string]int{"spe": 1}}
	out := strings.Join(bv.renderStatTable(p, true), "\n")
	if !strings.Contains(out, "151 +1") || !strings.Contains(out, "랭크 반영") || !strings.Contains(out, "미포함") {
		t.Fatalf("misleading modified stats: %s", out)
	}
}
