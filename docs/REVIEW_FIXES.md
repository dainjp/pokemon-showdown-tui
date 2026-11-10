# 기능 리뷰 수정 내역

작업 기준: `19c2e39` 이후 작업 트리. 1번과 3번은 기준 커밋에 이미 구현되어
있어 기존 로컬 WebSocket 회귀 테스트로 검증하고 유지했다.

| 리뷰 항목 | 반영한 동작 | 주요 코드/검증 |
| --- | --- | --- |
| 1. 도전 상태 | `/challenge` PM으로 받은/보낸 도전 및 취소 상태 갱신 | `showdown/parser.go`, `tui/challenges_test.go` |
| 2. 더블 선택 | 모든 활성 슬롯 선택 후 전송, 동시 강제 교체, 중복 교체 방지, 빈 슬롯 pass | `tui/battle.go`, `battle_actions.go`, `review_fixes_test.go` |
| 3. 팀 업로드 | 검색·도전·수락 전에 팀 업로드, 랜덤 규칙은 null | `tui/challenges.go`, `challenges_test.go` |
| 4. 관전 URL | 공식 play/replay URL과 방 ID 정규화, 쿼리/fragment 처리, 잘못된 주소 거부 | `tui/palette.go`, `TestBattleURLs` |
| 5. CLI 옵션 | 위치 인수 앞뒤 옵션 처리, doctor의 boolean sprites 유지 | `cmd/slowdown/main.go`, `main_test.go` |
| 6. 로그인 오류 | 성공 시 이전 오류 제거, 로그인 기억 설정 복원 | `tui/app.go`, `TestLoginSuccessAndRemember` |
| 7. 닫기/재접속 | leave와 재접속 목록 제거, 늦게 오는 닫힌 방 이벤트 무시, init 시 기록 재구축, 포커스 유지 | `tui/app.go`, `TestCloseSendsLeaveAndReplayResetsWithoutFocusSteal` |
| 8. 리플레이 | V/메뉴로 savereplay, 성공 응답의 실제 URL만 표시 | `tui/battle_actions.go`, `app.go` |
| 9. 전송 실패 | 전송 중/전송 완료 구분, 오류 표시와 재선택, 오래된 완료 이벤트 무시 | `tui/battle.go`, `app.go`, `TestFaintedMoveSlotAndSendFailure` |
| 10. 대타출동 | upstream 8d0b958의 SUB 배지·대타 스프라이트·로그와 테스트 적용, 로그 한국어 번역 | `battle/reducer.go`, `state.go`, `testdata/battles/substitute.txt` |
| 11. 팀 관리 | 이름/규칙/파일 내보내기/삭제 확인/선택 팀 대전, 실패한 저장의 메모리 변경 복구 | `tui/team_actions.go`, `teams/teams.go` |
| 12. 편의 기능 | Ctrl+T 타이머, U draft/전송 선택 취소, 친선 상대 재도전, 래더는 재검색 | `tui/battle_actions.go`, `review_fixes_test.go` |
| 13. 설정 | F2 기본 규칙/알림, keybindings/no_color/sidecar/sprites.size 연결 | `tui/settings.go`, `app.go`, `lobby.go`, `battle_render.go` |
| 기타 | 메뉴와 Q 종료 확인 통일, 로비 Q 종료, 도움말 F2/X 추가 | `tui/lobby.go`, `palette.go`, `app.go` |

팀 관리 검증 중 발견한 가져오기/내보내기 오류도 수정했다. 닉네임과 종의 순서,
성별 표기, 생략한 IV=31과 명시한 IV=0, packed 팀의 마지막 필드 및 테라타입을
보존한다. 기존 저장 데이터의 IV 전체 0/미지정 해석은 기본 31로 유지하며,
새로 명시한 전체 0 IV는 별도 표시로 구분한다. 이전 버전에서 이미 잘못 저장된
닉네임이나 부분 IV 정보는 추측해서 수정하지 않으며 원본 텍스트 재가져오기가 필요하다.

프로토콜 근거: 리플레이는 [공식 서버 방 구현](https://github.com/smogon/pokemon-showdown/blob/master/server/rooms.ts)의
uploadReplay 성공 응답을 사용한다. 팀 형식은 [공식 팀 변환 구현](https://github.com/smogon/pokemon-showdown/blob/master/sim/teams.ts)을 확인했다.

검증은 로컬 테스트 서버 및 저장소의 테스트를 사용한다. 실제 두 계정 사이의
공식 서버 대전과 Windows/macOS 실행은 이 작업에서 검증하지 않았다.

최종 확인 명령: `go test ./... -timeout 300s`, `go test -race ./... -timeout 300s`,
`go vet ./...`, `go build ./...`. Linux 실행 파일은 저장소 루트의 `slowdown`에
빌드하며 `--version`, `play gen9ou --version`, `doctor`로 CLI 실행을 확인한다.
