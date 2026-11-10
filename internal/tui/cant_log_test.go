package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/unnipv/pokemon-slowdown/internal/config"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

func TestPriorityBlockingAbilityLog(t *testing.T) {
	for _, tc := range []struct{ ability, korean string }{
		{"Queenly Majesty", "여왕의위엄"}, {"Dazzling", "비비드바디"}, {"Armor Tail", "테일아머"}, {"Damp", "습기"},
	} {
		t.Run(tc.ability, func(t *testing.T) {
			m := New(config.Default(), Deps{})
			bv := m.battleFor("battle-test")
			bv.state().Me = "p1"
			raw := "|switch|p1a: Donphan|Donphan, L100|100/100\n|switch|p2a: Tsareena|Tsareena, L100|100/100\n|cant|p2a: Tsareena|ability: " + tc.ability + "|Ice Shard|[of] p1a: Donphan\n|move|p2a: Tsareena|Power Whip|p1a: Donphan"
			for _, frame := range showdown.SplitFrames(raw) {
				for _, ev := range showdown.Parse(frame) {
					bv.apply(ev)
				}
			}
			log := bv.state().Log
			got := bv.logText(log[len(log)-2])
			want := "상대 달코퀸의 " + tc.korean + " 때문에 코리갑의 얼음뭉치 사용이 막혔습니다!"
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
			if got := bv.logText(log[len(log)-1]); got != "상대 달코퀸의 파워휩!" {
				t.Fatalf("following move: %q", got)
			}
			if bv.state().Find("p2a: Tsareena").Ability != tc.ability {
				t.Fatal("revealed ability not recorded")
			}
			m.cfg.Language = "en"
			if got := bv.logText(log[len(log)-2]); got != "The opposing Tsareena's "+tc.ability+" blocked Donphan's Ice Shard!" {
				t.Fatalf("English roles wrong: %q", got)
			}
		})
	}
}

func TestCantEffectsAndMissingAttacker(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-test")
	for _, tc := range []struct{ reason, translated string }{
		{"ability: Truant", "특성: 게으름"}, {"move: Taunt", "기술: 도발"}, {"item: Assault Vest", "도구: 돌격조끼"}, {"par", "마비"}, {"ability: Future Ability", "특성: Future Ability"},
	} {
		bv.apply(showdown.BattleCant{User: "p1a: Donphan", Reason: tc.reason, Move: "Ice Shard"})
		got := bv.logText(bv.state().Log[len(bv.state().Log)-1])
		if !strings.Contains(got, "코리갑의 얼음뭉치 사용 불가!") || !strings.Contains(got, tc.translated) || !utf8.ValidString(got) {
			t.Fatalf("reason %s: %q", tc.reason, got)
		}
	}
	bv.apply(showdown.BattleCant{User: "p2a: Tsareena", Reason: "ability: Queenly Majesty", Move: "Ice Shard"})
	got := bv.logText(bv.state().Log[len(bv.state().Log)-1])
	if got != "달코퀸의 여왕의위엄 때문에 얼음뭉치 사용이 막혔습니다!" {
		t.Fatalf("missing [of] invented attacker: %q", got)
	}
}

func TestPriorityBlockUsesTaggedAttackerInDoubles(t *testing.T) {
	m := New(config.Default(), Deps{})
	bv := m.battleFor("battle-test")
	bv.state().Me = "p1"
	raw := "|switch|p1a: Donphan|Donphan|100/100\n|switch|p1b: My Partner|Pikachu|100/100\n|switch|p2a: Tsareena|Tsareena|100/100\n|cant|p2a: Tsareena|ability: Queenly Majesty|Quick Attack|[of] p1b: My Partner"
	for _, frame := range showdown.SplitFrames(raw) {
		for _, ev := range showdown.Parse(frame) {
			bv.apply(ev)
		}
	}
	got := bv.logText(bv.state().Log[len(bv.state().Log)-1])
	if !strings.Contains(got, "My Partner의 전광석화") || strings.Contains(got, "코리갑") {
		t.Fatalf("wrong attacker/nickname: %q", got)
	}
}
