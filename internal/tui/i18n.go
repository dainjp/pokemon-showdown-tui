package tui

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/unnipv/pokemon-slowdown/internal/battle"
	"github.com/unnipv/pokemon-slowdown/internal/dex"
	"github.com/unnipv/pokemon-slowdown/internal/showdown"
)

//go:embed names_ko.json
var koreanNameData []byte

var koreanNames = func() map[string]map[string]string {
	var names map[string]map[string]string
	if err := json.Unmarshal(koreanNameData, &names); err != nil {
		panic("invalid embedded Korean catalogue: " + err.Error())
	}
	return names
}()

func (m *Model) tr(english string) string {
	if m.cfg.Language == "ko" {
		if korean, ok := koreanUI[english]; ok {
			return korean
		}
	}
	return english
}

func (bv *battleView) tr(english string) string {
	if bv.owner != nil {
		return bv.owner.tr(english)
	}
	return english
}

func (m *Model) name(category, english string) string {
	if m.cfg.Language != "ko" {
		return english
	}
	if korean := koreanNames[category][showdown.ToID(english)]; korean != "" {
		return korean
	}
	if category == "pokemon" {
		// Keep the form explicit when the catalogue only names the base species.
		for split := strings.LastIndex(english, "-"); split > 0; split = strings.LastIndex(english[:split], "-") {
			if base := koreanNames[category][showdown.ToID(english[:split])]; base != "" {
				form := english[split+1:]
				if korean := koreanForms[form]; korean != "" {
					form = korean
				}
				return base + " (" + form + ")"
			}
		}
	}
	return english
}

var koreanForms = map[string]string{
	"Therian": "영물폼", "Incarnate": "화신폼", "Origin": "오리진폼",
	"Mega": "메가", "Mega-X": "메가X", "Mega-Y": "메가Y",
	"Alola": "알로라", "Galar": "가라르", "Hisui": "히스이", "Paldea": "팔데아",
	"Wash": "워시폼", "Heat": "히트폼", "Frost": "프로스트폼", "Fan": "스핀폼", "Mow": "커트폼",
}

func (bv *battleView) name(category, english string) string {
	if bv.owner != nil {
		return bv.owner.name(category, english)
	}
	return english
}

func (bv *battleView) pokemonName(p *battle.Pokemon) string {
	if p == nil {
		return ""
	}
	// A nickname belongs to its player. Only translate an actual species name.
	if dex.ToID(p.Name) == dex.ToID(p.Species) {
		return SanitizeLine(bv.name("pokemon", p.Name))
	}
	if bv.deps.Dex != nil {
		if species, ok := bv.deps.Dex.Species(p.Species); ok && dex.ToID(p.Name) == dex.ToID(species.BaseSpecies) {
			return SanitizeLine(bv.name("pokemon", p.Name))
		}
	}
	return SanitizeLine(p.Name)
}

var generationLabel = regexp.MustCompile(`Gen ([1-9])`)
var koreanFormats = strings.NewReplacer(
	"Unrated Random Battle", "비랭크 랜덤 배틀",
	"Random Doubles Battle", "랜덤 더블 배틀",
	"Random Doubles", "랜덤 더블 배틀",
	"Random Battle", "랜덤 배틀",
	"Free-For-All", "배틀로열",
	"Multi", "멀티",
	"Custom Game", "자유 대전",
	"Doubles", "더블",
	"Little Cup", "리틀컵",
	"Monotype", "단일 타입",
	"Anything Goes", "무제한",
	"National Dex", "전국도감",
	"Champions", "챔피언스",
	"Blitz", "속전속결",
	"Battle", "배틀",
)

func (m *Model) formatLabel(f showdown.Format) string {
	label := formatLabel(f)
	if m.cfg.Language == "ko" {
		label = koreanFormats.Replace(generationLabel.ReplaceAllString(label, "${1}세대"))
	}
	return label
}

var koreanUI = map[string]string{
	"  [enter] rematch ":       "  [enter] 재대결: ",
	"Sending choice…":          "선택 전송 중…",
	"Could not send command: ": "명령 전송 실패: ",
	"Could not send choice: ":  "선택 전송 실패: ",
	"  V: save replay":         "  V: 리플레이 저장",
	"Save replay":              "리플레이 저장",
	"Enable battle timer":      "배틀 타이머 켜기",
	"Undo choice":              "선택 되돌리기",
	"Battle with this team":    "이 팀으로 대전",
	"Rename team":              "팀 이름 변경",
	"Set team format":          "팀 규칙 지정",
	"Export team":              "팀 내보내기",
	"Delete team":              "팀 삭제",
	"Name":                     "이름",
	"Format":                   "규칙",
	"New file path":            "새 파일 경로",
	"Delete team %s?":          "%s 팀을 삭제할까요?",
	"Set a format, then press Enter to battle.": "규칙을 지정한 뒤 Enter로 대전하세요.",
	"Team deleted.":                   "팀을 삭제했습니다.",
	"Enter a value.":                  "값을 입력하세요.",
	"Enter a format ID, e.g. gen9ou.": "규칙 ID를 입력하세요. 예: gen9ou",
	"Could not export team: ":         "팀 내보내기 실패: ",
	"Exported team to ":               "팀 저장 경로: ",
	"Team saved.":                     "팀을 저장했습니다.",
	"Enter battle · R rename · F format · E export · D delete · Esc back": "Enter 대전 · R 이름 · F 규칙 · E 내보내기 · D 삭제 · Esc 뒤로",
	"Default format":                                     "기본 규칙",
	"Notifications":                                      "알림",
	"  F2         settings":                              "  F2         설정",
	"  x          close finished battle":                 "  x          끝난 배틀 닫기",
	"  u          undo choice":                           "  u          선택 되돌리기",
	"  ctrl+t     enable battle timer":                   "  ctrl+t     배틀 타이머 켜기",
	"  v          save replay":                           "  v          리플레이 저장",
	"  enter      next ladder battle / friendly rematch": "  enter      다음 래더 대전 / 같은 상대와 친선 재대결",
	"F2 settings · : menu":                               "F2 설정 · : 메뉴",
	"Esc back · F2 close":                                "Esc 돌아가기 · F2 닫기",
	"Esc back":                                           "Esc 돌아가기",
	"1-4 moves · S switch · F2 settings · : menu":        "1–4 기술 · S 교체 · F2 설정 · : 메뉴",
	"1-4 moves · S switch · : menu":                      "1–4 기술 · S 교체 · : 메뉴",
	"1-4 move S swap : menu":                             "1–4 기술 S 교체 : 메뉴",
	"Status move":                                        "변화",
	"Tera":                                               "테라스탈",
	"Mega":                                               "메가진화",
	"Z-Move":                                             "Z기술",
	"Dynamax":                                            "다이맥스",
	"rain":                                               "비",
	"sandstorm":                                          "모래바람",
	"sun":                                                "쾌청",
	"hail":                                               "싸라기눈",
	"snow":                                               "눈",
	"harsh sunlight":                                     "강한 햇살",
	"heavy rain":                                         "강한 비",
	"strong winds":                                       "난기류",
	"It started to rain!":                                "비가 내리기 시작했습니다!",
	"A sandstorm kicked up!":                             "모래바람이 불기 시작했습니다!",
	"The sunlight got bright!":                           "햇살이 강해졌습니다!",
	"It started to hail!":                                "싸라기눈이 내리기 시작했습니다!",
	"It started to snow!":                                "눈이 내리기 시작했습니다!",
	"The sunlight turned extremely harsh!":               "햇살이 매우 강해졌습니다!",
	"A heavy rain began to fall!":                        "강한 비가 내리기 시작했습니다!",
	"Stat":                                               "능력치",
	"Base":                                               "종족값",
	"Actual":                                             "실능력치",
	"In battle":                                          "전투 중",
	"Min":                                                "최소",
	"Max":                                                "최대",
	"Attack":                                             "공격",
	"Defense":                                            "방어",
	"Sp. Atk":                                            "특수공격",
	"Sp. Def":                                            "특수방어",
	"Speed":                                              "스피드",
	"accuracy":                                           "명중률",
	"evasion":                                            "회피율",
	"burned":                                             "화상",
	"paralyzed":                                          "마비",
	"poisoned":                                           "독",
	"badly poisoned":                                     "맹독",
	"asleep":                                             "잠듦",
	"frozen":                                             "얼음",
	"rose":                                               "상승",
	"rose sharply":                                       "크게 상승",
	"rose drastically":                                   "매우 크게 상승",
	"fell":                                               "하락",
	"fell harshly":                                       "크게 하락",
	"fell severely":                                      "매우 크게 하락",
	"Connection closed after a proxy-lock notice. Automatic retries paused; use : to sign in again.": "프록시 잠금 안내 후 연결이 종료됐습니다. 자동 재시도 중지 · : 메뉴에서 다시 로그인",
	"Showdown reported a proxy lock for this network's IP. See pokemonshowdown.com/pages/proxyhelp.": "Showdown이 이 네트워크 IP의 프록시 잠금을 알렸습니다. 안내: pokemonshowdown.com/pages/proxyhelp",
	"Settings":                         "설정",
	"Language":                         "언어",
	"Sprites":                          "스프라이트",
	"Animation":                        "애니메이션",
	"Theme":                            "테마",
	"On":                               "켜짐",
	"Off":                              "꺼짐",
	"Changes are saved automatically.": "바꾸는 즉시 자동으로 저장합니다.",
	"↑/↓ select · Enter change · Esc back": "↑/↓ 선택 · Enter 변경 · Esc 돌아가기",
	"F2 settings · ? help":                 "F2 설정 · ? 도움말",
	"F2 to change settings":                "F2를 눌러 설정 변경",
	"Could not change sprites: ":           "스프라이트를 바꾸지 못했습니다: ",
	"Searching ":                           "대전 검색 중: ",
	"random":                               "랜덤",
	"challenge":                            "친선",
	"online":                               "연결됨",
	"offline":                              "연결 종료",
	"connecting":                           "접속 중",
	"reconnecting":                         "재접속 중",
	"Find a battle":                        "대전 찾기",
	"search formats…":                      "규칙 이름을 검색하세요…",
	"  no formats matched\n":               "  검색 결과가 없습니다\n",
	"Status":                               "상태",
	"Battles":                              "대전 목록",
	"Challenges":                           "받은 도전",
	"  : cancel search":                    "  : 검색 취소",
	"  : accept / reject":                  "  : 수락 / 거절",
	"Challenging ":                         "도전 중: ",
	"connection: ":                         "연결: ",
	"not signed in as ":                    "로그인 필요: ",
	"  : sign in":                          "  : 로그인",
	"guest · : sign in for ladder ranking": "게스트 · : 메뉴에서 로그인",
	"enter queue · ctrl+t teams · tab battle · : palette · ? help": "Enter 대전 · Ctrl+T 팀 · F2 설정 · : 메뉴",
	"Your teams":                      "내 팀",
	"  No teams stored yet.":          "  아직 저장한 팀이 없습니다.",
	"  Import one with : import team": "  : 메뉴 → 팀 가져오기에서 추가하세요.",
	"Preview":                         "미리보기",
	"esc back · : palette to import or export":                    "Esc 돌아가기 · : 메뉴에서 팀 가져오기",
	"Quit pokemon slowdown?":                                      "앱을 종료할까요?",
	"No battle selected. Press esc to return to the lobby.":       "선택한 대전이 없습니다. Esc로 로비에 돌아가세요.",
	"Terminal too small\n%d×%d (need %d×%d)\nResize to continue.": "터미널 창이 작습니다\n%d×%d (최소 %d×%d)\n창 크기를 늘려주세요.",
	"esc lobby": "Esc 로비 · F2 설정",
	"1-4 move  s switch  i inspect  l log  c chat  tab battle  : palette": "1–4 기술  S 교체  I 정보  L 기록  C 채팅  F2 설정  : 메뉴",
	"1-4 move  s switch  tab battle  : palette":                           "1–4 기술  S 교체  Tab 대전  : 메뉴",
	"esc close":                  "Esc 닫기",
	"guest":                      "게스트",
	"searching ":                 "대전 검색 중: ",
	"keyboard":                   "키보드 조작",
	"  1-4        choose a move": "  1–4        기술 선택",
	"  s          switch (then 1-6, or a name)":                             "  S          포켓몬 교체 (1–6 선택)",
	"  t          use the current mechanic (Tera/Mega/Z/Dynamax)":           "  T          테라스탈 / 메가진화 / Z기술 / 다이맥스",
	"  i          inspect a Pokémon (↑/↓ cycles both sides and your party)": "  I          포켓몬 정보 (↑/↓로 선택)",
	"  l          battle log":                                               "  L          전투 기록",
	"  c          battle chat":                                              "  C          대전 채팅",
	"  tab        switch between open battles":                              "  Tab        진행 중인 대전 전환",
	"  enter      (after a battle) queue the same format again":             "  Enter      대전 종료 후 같은 규칙으로 다시 검색",
	"  :          command palette":                                          "  :          전체 메뉴",
	"  ?          this help":                                                "  ?          도움말",
	"  esc        close / back to lobby":                                    "  Esc        닫기 / 로비로 돌아가기",
	"  q          quit (asks first during a battle)":                        "  Q          종료 (대전 중에는 확인)",
	"  ctrl+c     quit immediately":                                         "  Ctrl+C     바로 종료",
	"party tracker":                                                         "파티 표시",
	" seen and alive     ":                                                  " 확인됨 · 생존     ",
	" still hidden":                                                         " 아직 미공개",
	" currently active   ":                                                  " 현재 출전        ",
	" fainted":                                                              " 기절",
	"Taglines: ":                                                            "한마디: ",
	"enter / y to confirm   esc / n to cancel":                              "Enter / Y 확인   Esc / N 취소",
	"Queue ":                           "대전 검색: ",
	"Cancel search":                    "대전 검색 취소",
	"Search cancelled.":                "대전 검색을 취소했습니다.",
	"Find a battle (format picker)":    "대전 찾기 · 규칙 선택",
	"Sign out of ":                     "로그아웃: ",
	"Sign in to a registered account…": "Showdown 계정으로 로그인…",
	"Sign in to Pokémon Showdown":      "Showdown 로그인",
	"Username":                         "아이디",
	"Password":                         "비밀번호",
	"Remember":                         "로그인 기억",
	"Challenge a user…":                "다른 사용자에게 도전…",
	"Challenge a user":                 "사용자에게 도전",
	"Challenge format: ":               "도전 규칙 선택: ",
	"Cancel challenge to ":             "보낸 도전 취소: ",
	"  : cancel challenge":             "  : 메뉴 → 보낸 도전 취소",
	"Accepting challenge from ":        "도전 수락 요청 중: ",
	"Challenge from ":                  "받은 도전: ",
	"Import a team first: : menu → Import a team.": "팀 가져오기가 필요합니다. : 메뉴 → 팀 가져오기",
	"Could not send battle request: ":              "대전 요청을 보내지 못했습니다: ",
	"Enter challenge · Esc cancel · : menu":        "Enter 도전 · Esc 취소 · : 메뉴",
	"Enter challenge · Esc cancel":                 "Enter 도전 · Esc 취소",
	"Accept challenge from ":                       "도전 수락: ",
	"Reject challenge from ":                       "도전 거절: ",
	"Accepted challenge from ":                     "도전을 수락했습니다: ",
	"Open battle: ":                                "대전 열기: ",
	"Teams":                                        "내 팀",
	"Import team from clipboard text…":             "팀 가져오기 · 내보내기 텍스트 붙여넣기…",
	"Import a team":                                "팀 가져오기",
	"Team text":                                    "팀 텍스트",
	"Spectate a battle…":                           "대전 관전…",
	"Spectate a battle":                            "대전 관전",
	"Battle id":                                    "대전 ID",
	"Toggle sprites (currently ":                   "스프라이트 켜기 / 끄기 (현재 ",
	"Toggle animation (currently %v)":              "애니메이션 켜기 / 끄기 (현재 %v)",
	"Sprites: ":                                    "스프라이트: ",
	"Animation: %v":                                "애니메이션: %v",
	"Next theme":                                   "다음 테마",
	"Theme: ":                                      "테마: ",
	"Save settings":                                "설정 저장",
	"Could not save settings: ":                    "설정을 저장하지 못했습니다: ",
	"Settings saved.":                              "설정을 저장했습니다.",
	"Close this finished battle":                   "종료한 대전 닫기",
	"Close all finished battles":                   "종료한 대전 모두 닫기",
	"Forfeit this battle":                          "이 대전 기권",
	"Forfeit this battle?":                         "이 대전을 기권할까요?",
	"Keyboard help":                                "키보드 도움말",
	"Quit":                                         "종료",
	"Commands":                                     "전체 메뉴",
	"type a command…":                              "메뉴 이름을 검색하세요…",
	"  no matching command":                        "  일치하는 메뉴가 없습니다",
	"↑/↓ choose · enter run · esc close":           "↑/↓ 선택 · Enter 실행 · Esc 닫기",
	"Enter a username.":                            "아이디를 입력하세요.",
	"Enter your password. (Guest names are assigned by the server.)":                           "비밀번호를 입력하세요. 게스트 이름은 서버가 정합니다.",
	"Could not store the password in your keychain; you will need to sign in again next time.": "키체인에 저장하지 못했습니다. 다음 실행 때 다시 로그인하세요.",
	"Signing in as ":                        "로그인 중: ",
	"Could not parse team: ":                "팀을 읽을 수 없습니다: ",
	"Team storage unavailable.":             "팀 저장소를 사용할 수 없습니다.",
	"Could not save team: ":                 "팀을 저장하지 못했습니다: ",
	"Imported %d Pokémon.":                  "포켓몬 %d마리를 가져왔습니다.",
	"Could not read a battle id from that.": "대전 ID 또는 리플레이 주소를 확인하세요.",
	"Joining ":                              "입장 중: ",
	"Signed out.":                           "로그아웃했습니다.",
	"enter next · esc cancel":               "Enter 다음 · Tab 이동 · Esc 취소",
	"enter confirm · esc cancel":            "Enter 확인 · Esc 취소",
	"Could not load dex data: ":             "도감 데이터를 불러오지 못했습니다: ",
	"Reconnected.":                          "다시 연결했습니다.",
	"Login failed: ":                        "로그인 실패: ",
	"Logged in as ":                         "로그인 완료: ",
	"Your turn!":                            "내 차례입니다!",
	"Turn ":                                 "턴 ",
	"Your turn in ":                         "내 차례: ",
	"Closed ":                               "닫았습니다: ",
	"Battle closed.":                        "대전을 닫았습니다.",
	"No finished battles to close.":         "닫을 대전이 없습니다.",
	"Closed %d finished battle(s).":         "종료한 대전 %d개를 닫았습니다.",
	"Turn %d":                               "%d턴",
	"opponent · ":                           "상대 · ",
	"  (no active Pokémon)":                 "  (출전한 포켓몬 없음)",
	"  fainted":                             "  기절",
	"you":                                   "나",
	"foe":                                   "상대",
	"ally":                                  "아군",
	"tera:":                                 "테라:",
	"  You must switch in a Pokémon.  ":     "  교체할 포켓몬을 선택하세요.  ",
	"press s or 1-6":                        "S 또는 1–6",
	"  choosing for slot %d of %d":          "  행동 선택: %d / %d",
	"choice sent":                           "선택 완료",
	"— waiting for the opponent…":           "· 상대를 기다리는 중…",
	"chosen: ":                              "선택: ",
	"mechanic: ":                            "특수 행동: ",
	"[s] switch":                            "[S] 교체",
	"Phys":                                  "물리",
	"Spec":                                  "특수",
	"  Choose your lead order":              "  선출 순서를 정하세요",
	"  press a number to bring a Pokémon forward, enter to confirm": "  숫자로 순서 선택 · Enter 확정",
	"  The battle ended in a tie.":                                  "  무승부입니다.",
	" won the battle!":                                              " 승리!",
	"  replay: https://replay.pokemonshowdown.com/":                 "  리플레이: https://replay.pokemonshowdown.com/",
	"  [enter] queue ":                                              "  [Enter] 다시 검색: ",
	" again":                                                        "",
	"  [esc] lobby   [tab] switch battle   [x] close   [l] log   [c] chat": "  [Esc] 로비  [Tab] 다른 대전  [X] 닫기  [L] 기록  [C] 채팅",
	"◌ waiting":                   "대기 중",
	" · waiting":                  " · 대기 중",
	"… %d more lines":             "… %d줄 더 있음",
	"Switch":                      "포켓몬 교체",
	"No party information yet.":   "아직 파티 정보가 없습니다.",
	"unavailable":                 "선택 불가",
	"fainted":                     "기절",
	"in battle":                   "출전 중",
	"enter choose  ·  i inspect":  "Enter 선택 · I 정보",
	"Inspect":                     "포켓몬 정보",
	"Nothing to inspect.":         "확인할 포켓몬이 없습니다.",
	"  %d of %d  ·  ↑/↓ to cycle": "  %d / %d · ↑/↓ 다른 포켓몬",
	"Status   %s":                 "상태     %s",
	"Level    %d":                 "레벨     %d",
	"Tera     ":                   "테라     ",
	"Boosts   ":                   "능력 변화 ",
	"Item     ":                   "도구     ",
	"Ability  ":                   "특성     ",
	"? (not revealed)":            "? (미공개)",
	"Moves":                       "기술",
	"disabled":                    "사용 불가",
	"  PP shown while active":     "  출전 중인 포켓몬의 PP를 표시합니다",
	"  (not yet known)":           "  (아직 미공개)",
	"Revealed moves (%d)":         "공개된 기술 (%d개)",
	"  none seen yet":             "  아직 확인한 기술이 없습니다",
	"Only what the battle has revealed is shown.":           "전투 중 공개된 정보만 표시합니다.",
	"  Possible range, before items, abilities and boosts.": "  도구·특성·능력 변화를 적용하기 전의 범위입니다.",
	"(nothing yet)":     "(아직 기록이 없습니다)",
	"↑/↓ scroll":        "↑/↓ 스크롤",
	"Battle log":        "전투 기록",
	"(no messages)":     "(메시지가 없습니다)",
	"Chat":              "채팅",
	"Choose a target":   "대상 선택",
	"What will you do?": "무엇을 할까요?",
	"Somewhere, your opponent is staring at a screen.": "상대가 행동을 선택하고 있습니다.",
	"Please wait.":                           "잠시 기다려주세요.",
	"Congratulations! You won!":              "축하합니다! 승리했어요!",
	"You whited out!":                        "눈앞이 깜깜해졌다!",
	"The Pokémon Center is a place to rest.": "다음 대전을 준비해보세요.",
}
