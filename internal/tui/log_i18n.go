package tui

import (
	"fmt"
	"strings"

	"github.com/unnipv/pokemon-slowdown/internal/battle"
)

type logTranslation struct {
	text string
	// Categories identify game data; an empty category leaves player names alone.
	categories []string
}

func (bv *battleView) logText(entry battle.LogEntry) string {
	if bv.owner == nil || bv.owner.cfg.Language != "ko" {
		return entry.Text
	}
	translation, ok := koreanLog[entry.Template]
	if !ok {
		return bv.tr(entry.Text)
	}
	args := append([]any(nil), entry.Args...)
	for i, category := range translation.categories {
		if i >= len(args) {
			break
		}
		text, ok := args[i].(string)
		if !ok || category == "" {
			continue
		}
		if category == "pokemon" {
			prefix := ""
			if strings.HasPrefix(text, "The opposing ") {
				prefix = "상대 "
				text = strings.TrimPrefix(text, "The opposing ")
			}
			translated := bv.name("pokemon", text)
			for _, side := range []*battle.Side{bv.state().P1, bv.state().P2} {
				if side == nil {
					continue
				}
				for _, pokemon := range side.Party {
					if pokemon.Name == text {
						translated = bv.pokemonName(pokemon)
					}
				}
			}
			args[i] = prefix + translated
		} else if category == "text" {
			args[i] = bv.tr(text)
		} else {
			args[i] = bv.name(category, text)
		}
	}
	return fmt.Sprintf(translation.text, args...)
}

var koreanLog = map[string]logTranslation{
	"%s made a substitute!":         {"%s의 대타가 나타났다!", []string{"pokemon"}},
	"%s's substitute faded!":        {"%s의 대타가 사라졌다!", []string{"pokemon"}},
	"%s's substitute took the hit!": {"%s의 대타가 공격을 받았다!", []string{"pokemon"}},
	"The substitute took the hit!":  {"대타가 공격을 받았다!", nil},
	"— Turn %d —":                   {"— %d턴 —", nil},
	"%s won the battle!":            {"%s 승리!", nil},
	"The battle ended in a tie.":    {"무승부입니다.", nil},
	"Battle timer: %s":              {"대전 타이머: %s", nil},
	"Battle timer off.":             {"대전 타이머를 껐습니다.", nil},
	"%s used %s!":                   {"%s의 %s!", []string{"pokemon", "move"}},
	"%s changed forme to %s!":       {"%s의 모습이 %s(으)로 바뀌었습니다!", []string{"pokemon", "pokemon"}},
	"%s was revealed to be %s!":     {"%s의 정체는 %s!", []string{"pokemon", "pokemon"}},
	"%s can't use %s! (%s)":         {"%s은(는) %s을(를) 쓸 수 없습니다! (%s)", []string{"pokemon", "move", "text"}},
	"%s can't move! (%s)":           {"%s은(는) 움직일 수 없습니다! (%s)", []string{"pokemon", "text"}},
	"%s fainted!":                   {"%s은(는) 쓰러졌습니다!", []string{"pokemon"}},
	"%s lost %s%s":                  {"%s의 HP 감소: %s%s", []string{"pokemon"}},
	"%s restored %s%s":              {"%s의 HP 회복: %s%s", []string{"pokemon"}},
	"%s was %s!":                    {"%s의 상태: %s!", []string{"pokemon", "text"}},
	"%s recovered from %s.":         {"%s의 %s 상태가 나았습니다.", []string{"pokemon", "text"}},
	"A bell chimed and cured the team's status conditions.": {"방울 소리로 팀의 상태이상이 나았습니다.", nil},
	"%s's %s %s!":                               {"%s의 %s: %s!", []string{"pokemon", "text", "text"}},
	"%s swapped stat changes with %s.":          {"%s과(와) %s의 능력 변화를 바꿨습니다.", []string{"pokemon", "pokemon"}},
	"%s's stat changes were inverted!":          {"%s의 능력 변화가 반대가 됐습니다!", []string{"pokemon"}},
	"%s's stat changes were cleared!":           {"%s의 능력 변화가 사라졌습니다!", []string{"pokemon"}},
	"All stat changes were eliminated!":         {"모든 능력 변화가 사라졌습니다!", nil},
	"%s copied %s's stat changes!":              {"%s은(는) %s의 능력 변화를 복사했습니다!", []string{"pokemon", "pokemon"}},
	"The weather returned to normal.":           {"날씨가 원래대로 돌아왔습니다.", nil},
	"%s activated.":                             {"%s 발동!", []string{"move"}},
	"%s started on %s's side.":                  {"%s: %s 쪽에 적용됐습니다.", []string{"move", ""}},
	"%s ended on %s's side.":                    {"%s: %s 쪽에서 사라졌습니다.", []string{"move", ""}},
	"Side conditions were swapped!":             {"양쪽의 설치 효과가 바뀌었습니다!", nil},
	"%s revealed its %s!":                       {"%s의 도구는 %s!", []string{"pokemon", "item"}},
	"%s ate its %s!":                            {"%s은(는) %s을(를) 먹었습니다!", []string{"pokemon", "item"}},
	"%s lost its %s!":                           {"%s은(는) %s을(를) 잃었습니다!", []string{"pokemon", "item"}},
	"%s's %s!":                                  {"%s의 %s!", []string{"pokemon", "ability"}},
	"%s's ability became %s!":                   {"%s의 특성이 %s(으)로 바뀌었습니다!", []string{"pokemon", "ability"}},
	"%s's ability was suppressed!":              {"%s의 특성이 무효화됐습니다!", []string{"pokemon"}},
	"%s transformed into %s!":                   {"%s은(는) %s(으)로 변신했습니다!", []string{"pokemon", "pokemon"}},
	"%s Mega Evolved!":                          {"%s의 메가진화!", []string{"pokemon"}},
	"%s underwent Primal Reversion!":            {"%s의 원시회귀!", []string{"pokemon"}},
	"%s Ultra Burst!":                           {"%s의 울트라버스트!", []string{"pokemon"}},
	"%s unleashed its Z-Power!":                 {"%s은(는) Z파워를 해방했습니다!", []string{"pokemon"}},
	"The Z-Move broke through %s's protection!": {"Z기술이 %s의 방어를 뚫었습니다!", []string{"pokemon"}},
	"A critical hit!":                           {"급소에 맞았습니다!", nil},
	"It's super effective!":                     {"효과가 굉장했다!", nil},
	"It's not very effective...":                {"효과가 별로인 듯하다…", nil},
	"It doesn't affect %s...":                   {"%s에게는 효과가 없습니다…", []string{"pokemon"}},
	"But %s's %s failed!":                       {"하지만 %s의 %s은(는) 실패했습니다!", []string{"pokemon", "move"}},
	"But it failed!":                            {"하지만 실패했습니다!", nil},
	"%s's attack missed!":                       {"%s의 공격은 빗나갔습니다!", []string{"pokemon"}},
	"%s Terastallized into the %s type!":        {"%s은(는) %s 타입으로 테라스탈했습니다!", []string{"pokemon", "type"}},
	"Hit %s time(s)!":                           {"%s회 맞았습니다!", nil},
	"%s is preparing %s!":                       {"%s은(는) %s을(를) 준비합니다!", []string{"pokemon", "move"}},
	"%s must recharge!":                         {"%s은(는) 반동으로 쉬고 있습니다!", []string{"pokemon"}},
	"But nothing happened!":                     {"하지만 아무 일도 일어나지 않았습니다!", nil},
	"There was no target...":                    {"대상이 없습니다…", nil},
	"%s was dragged out!":                       {"%s이(가) 끌려 나왔습니다!", []string{"pokemon"}},
	"%s was sent out!":                          {"%s이(가) 나왔습니다!", []string{"pokemon"}},
	"Go! %s!":                                   {"가랏! %s!", []string{"pokemon"}},
	"%s sent out %s!":                           {"%s은(는) %s을(를) 내보냈습니다!", []string{"", "pokemon"}},
}
