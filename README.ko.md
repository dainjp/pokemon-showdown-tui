# Pokémon Showdown TUI 설치 가이드

기존 Pokémon Showdown 서버에서 브라우저 사용자와 대전하는 터미널 클라이언트입니다.
[unnipv/pokemon-slowdown](https://github.com/unnipv/pokemon-slowdown)을 기반으로
한국어, 스프라이트 ON/OFF, F2 설정, `midnight` 테마를 추가했습니다.
실행 명령 이름은 `slowdown`입니다.

## 1. 내 PC용 파일 받기

[v0.3.0-ko.2 다운로드 페이지](https://github.com/dainjp/pokemon-showdown-tui/releases/tag/v0.3.0-ko.2)의
**Assets**에서 아래 파일을 받으세요. Go나 Git 설치는 필요하지 않습니다.
이 저장소는 비공개이므로 접근 권한이 있는 GitHub 계정으로 로그인해야 합니다.

| PC | 받을 파일 |
| --- | --- |
| Windows x64 (일반 Intel/AMD PC) | `pokemon-showdown-tui_0.3.0-ko.2_windows_amd64.zip` |
| Mac Apple Silicon (M1 이후) | `pokemon-showdown-tui_0.3.0-ko.2_darwin_arm64.tar.gz` |
| Mac Intel | `pokemon-showdown-tui_0.3.0-ko.2_darwin_amd64.tar.gz` |
| Linux x64 | `pokemon-showdown-tui_0.3.0-ko.2_linux_amd64.tar.gz` |
| Linux ARM64 | `pokemon-showdown-tui_0.3.0-ko.2_linux_arm64.tar.gz` |

Windows ARM64 전용 빌드는 제공하지 않습니다. `Source code` 파일은 개발용 소스이며
바로 실행하는 프로그램이 아닙니다. Mac 종류는 Apple 메뉴 → **이 Mac에 관하여**에서 확인합니다.

## 2. Windows에서 실행

1. 받은 ZIP을 **모두 압축 풀기**로 풀어주세요.
2. 압축을 푼 폴더를 열고 우클릭 → **터미널에서 열기**를 선택합니다.
3. 로컬 PowerShell에서 아래 명령을 입력합니다. Ubuntu에 SSH로 접속한 창에서는 실행하지 마세요.

```powershell
.\slowdown.exe --no-sprites
```

한국어 화면과 스프라이트 없는 모드로 시작합니다. 창이 계속 열려 있어 오류도 확인할 수 있습니다.
프로그램 실행이 차단되면 배포 페이지와 파일명을 확인한 후 표시되는 Windows 보안 안내를 따르세요.

## 3. Mac에서 실행

1. 받은 `.tar.gz` 파일을 압축 해제합니다.
2. 압축을 푼 폴더에서 Terminal을 엽니다. Finder의 폴더 우클릭 → **폴더에서 새로운 터미널 열기**를 사용할 수 있습니다.
3. 아래 명령을 실행합니다.

```sh
./slowdown --no-sprites
```

`Permission denied`이면 `chmod +x ./slowdown` 후 다시 실행하세요.
개발자 확인 때문에 차단되면 배포 페이지와 파일명을 확인한 뒤 macOS의
**시스템 설정 → 개인정보 보호 및 보안**에서 해당 실행 파일의 허용 안내를 따르세요.
이 테스트 빌드는 Apple 공증을 받지 않았습니다.

## 4. 로그인하고 기억하기

1. 앱에서 `:`를 누르고 **Showdown 계정으로 로그인…**을 선택합니다.
2. 기존 Showdown 아이디와 비밀번호를 입력합니다. Enter 또는 Tab으로 다음 칸으로 이동합니다.
3. **로그인 기억**에서 `Y`를 누른 후 Enter로 제출합니다. `N`은 끄기, Space는 전환입니다.

입력 칸은 Ctrl+U로 비울 수 있습니다. 비밀번호는 설정 파일에 쓰지 않고,
기억을 켰을 때 Windows 자격 증명 관리자 또는 macOS 키체인에 보관합니다.
키링에 저장할 수 없으면 앱에서 오류를 알리고 다음 실행 때 다시 입력해야 합니다.
Windows와 Mac의 설정·팀·로그인 정보는 각 PC에 따로 저장됩니다.

## 5. 화면 설정과 대전

`F2`에서 언어, 스프라이트, 애니메이션, 테마를 바꾸면 자동 저장됩니다.
Mac에서 F2가 밝기 조절로 동작하면 `fn + F2`를 사용하세요.

| 키 | 기능 |
| --- | --- |
| `F2`, `:` | 설정, 전체 메뉴 |
| `1`–`4`, `S`, `T` | 기술 선택, 교체, 테라스탈 등 특수 행동 |
| `I`, `L`, `C` | 포켓몬 정보, 전투 기록, 채팅 |
| `?`, `Esc`, `Ctrl+C` | 도움말, 돌아가기, 종료 |

로그인 후 로비에서 랜덤 배틀을 선택하면 팀 준비 없이 검색할 수 있습니다.
직접 만든 팀을 쓰려면 `:` 메뉴의 팀 가져오기에 Showdown 팀 내보내기 텍스트를 붙여넣으세요.
친구에게 도전하려면 `:` → **다른 사용자에게 도전…**에서 아이디를 입력하고,
이어지는 **도전 규칙 선택** 목록에서 검색 또는 ↑/↓로 규칙을 고른 뒤 Enter를 누릅니다.
기본 규칙 외에 OU·더블·커스텀 게임 등 서버가 도전을 허용하는 규칙을 선택할 수 있습니다.
Esc는 전송하지 않고 선택을 취소합니다. CLI의 `slowdown challenge <아이디>`도 같은 선택 화면을 엽니다.

보낸 신청을 취소하려면 `:` → **보낸 도전 취소: 상대 아이디**를 선택하세요.
서버의 도전 상태를 받아 수락·거절·취소를 반영하며, 여러 상대에게 보낸 신청도 각각 취소할 수 있습니다.
받은 도전의 수락·거절과 기존 배틀 관전도 `:` 메뉴에서 할 수 있습니다.

팀이 필요한 규칙은 도전·수락·검색 직전에 저장된 팀을 서버에 전송합니다.
규칙이 일치하는 팀을 우선 사용하고, 없으면 비어 있지 않은 첫 팀을 사용합니다.
저장한 팀이 없으면 **팀 가져오기가 필요합니다**라고 안내하고 요청을 보내지 않습니다.
랜덤 배틀은 저장된 팀을 쓰지 않습니다. 팀의 해당 규칙 적합성은 서버가 최종 검증합니다.

`--no-sprites`는 시작할 때 스프라이트를 끄는 옵션입니다. 다음 실행에도 끈 상태를
유지하려면 `F2`에서 스프라이트 설정을 변경해 저장하거나 매번 같은 옵션으로 실행하세요.
영어 화면은 `--lang en`, 스프라이트 자동 선택은 `--sprites auto`를 사용합니다.

## 프록시 잠금 / Reconnecting이 보이면

Ubuntu 개발 서버에서 **프록시 잠금 안내 → 인증 성공 → 연결 종료**가 관찰됐습니다.
클라이언트는 이 순서가 나오면 자동 재접속을 멈춰 로그인 반복을 줄입니다.
로그인이 유지되는 문제까지 해결했다고 보장하지 않으며, 잠금 안내 자체가 연결 종료의
직접 원인인지도 아직 확인하지 못했습니다.

PC에서 직접 실행하면 접속 네트워크를 바꿔 비교할 수 있습니다. 로컬에서도 잠금이 나오면
사용 중인 VPN·프록시 설정을 확인하고
[Pokémon Showdown 공식 안내](https://pokemonshowdown.com/pages/proxyhelp)를 따르세요.
프로그램이 서버의 잠금을 해제할 수는 없습니다.

오류를 확인할 때는 실행 명령, OS, 화면의 오류 문구를 알려주세요.
비밀번호나 인증 토큰은 공유하지 마세요. 터미널 자체 점검은 다음과 같습니다.

```powershell
# Windows
.\slowdown.exe doctor
```

```sh
# macOS / Linux
./slowdown doctor
```

## 소스를 받아 직접 수정하기

Git과 Go 1.27.1 이상이 필요합니다. 비공개 저장소를 복제할 수 있는 GitHub 인증을
먼저 설정하세요. 아래 명령은 로컬 개발 PC에서 실행합니다.

```sh
git clone https://github.com/dainjp/pokemon-showdown-tui.git
cd pokemon-showdown-tui
go test ./... -timeout 300s
go vet ./...
go build -o slowdown ./cmd/slowdown
```

Windows PowerShell에서는 마지막 빌드 명령을 바꿔주세요.

```powershell
go build -o slowdown.exe ./cmd/slowdown
.\slowdown.exe --no-sprites
```

macOS / Linux에서는 `./slowdown --no-sprites`로 실행합니다.
새 코드를 받을 때는 저장소 폴더에서 `git pull --ff-only` 후 다시 빌드하세요.
Go 모듈 경로는 원본과 호환되도록 유지했습니다. 원본 Homebrew/Go 설치 명령은
이 개인화 저장소의 코드를 설치하지 않으므로 위 clone/build 또는 배포 파일을 사용하세요.

| 위치 | 수정할 내용 |
| --- | --- |
| `internal/tui/i18n.go`, `log_i18n.go` | 메뉴·전투 기록 번역 |
| `internal/tui/settings.go` | 설정 화면과 저장 동작 |
| `internal/tui/theme.go` | 테마 색과 스타일 |
| `internal/tui/battle_render.go` | 전투 화면 배치 |
| `scripts/update-korean-names.py` | PokéAPI 한국어 이름 데이터 갱신 |

설정 파일은 기본적으로 `~/.config/pokemon-slowdown/config.toml`, Windows에서는
`%USERPROFILE%\.config\pokemon-slowdown\config.toml`입니다.
`XDG_CONFIG_HOME`이 설정돼 있으면 그 경로를 사용합니다.
포켓몬·기술·도구·특성·타입 이름과 주요 전투 기록을 한국어로 표시하며,
플레이어 닉네임·채팅·팀 텍스트·서버 명령은 원문을 유지합니다.
일부 폼·기술 설명·서버 안내·전투 문구는 영어로 남습니다.
라이선스와 데이터 출처는 `LICENSE`, `THIRD_PARTY.md`, `docs/POKEAPI-LICENSE.txt`를 참고하세요.

## 확인한 범위

Linux 실행, 게스트 서버 연결, F2 설정과 저장, 기록된 전투 렌더링, 단위 테스트와
경쟁 상태 검사, Windows/macOS 교차 빌드를 확인했습니다.
Windows/macOS의 실제 터미널 실행, 계정 로그인 유지, 실제 사람과의 대전은
로컬 PC에서 확인할 항목입니다. 이 버전은 첫 테스트 배포입니다.
