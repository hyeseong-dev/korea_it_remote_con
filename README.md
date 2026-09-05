# Portable Remote Access v2.0.0

장치 별칭과 로컬 프로필로 AnyDesk 연결을 여는 공통 CLI입니다. macOS·Windows·Linux의 앱 실행 차이를 어댑터로 분리합니다. 화면 데이터는 기존 AnyDesk가 처리하며, 이 도구는 중계 서버나 상시 실행 서비스가 아닙니다. RDP 및 원격 PC의 설정은 변경하지 않습니다.

## 요구 사항

- Python 3.9 이상과 로컬 AnyDesk 클라이언트
- 원격 장치의 AnyDesk 실행 및 사전에 설정한 무인 접속 인증
- 주 경로: 양쪽 Tailscale 연결 및 원격 AnyDesk TCP 포트 접근
- 비상 경로: AnyDesk 서비스 중계에 접근 가능한 인터넷 연결

외부 Python 패키지는 필요하지 않습니다. Windows에서는 아래 `python3` 대신 `py -3`를 사용할 수 있습니다. 현재 개발 Mac의 Python site 초기화 오류가 발생하면 `python3 -S`로 실행하세요. OS별 단독 실행 파일 패키징은 아직 제공하지 않습니다.

## 시작

프로젝트 폴더에서 예시를 복사합니다. 이미 설정이 있으면 덮어쓰지 마세요.

```sh
# macOS / Linux
cp -n examples/devices.json devices.local.json
chmod 600 devices.local.json
```

```powershell
# Windows PowerShell
if (-not (Test-Path devices.local.json)) {
    Copy-Item examples/devices.json devices.local.json
}
```

`devices.local.json`의 예시 주소를 실제 접속 주소로 수정합니다. 예시 주소는 실제 접속 대상이 아닙니다. 현재 작업 Mac에는 기존 로컬 바로가기에서 실제 설정을 옮겨 두었습니다.

```sh
python3 remote.py list
python3 remote.py connect academy --dry-run
python3 remote.py doctor academy
python3 remote.py connect academy
python3 remote.py connect academy --route fallback
```

다른 프로필 파일을 사용할 수 있습니다.

```sh
python3 remote.py --config /path/to/devices.local.json list
```

인증은 AnyDesk가 담당합니다. 이 도구는 암호를 읽거나 저장하지 않으며 로그인 창을 자동 조작하지 않습니다. `launched`는 앱에 실행 요청을 전달했다는 뜻이고, 원격 연결·인증 성공을 뜻하지 않습니다.

## 구조

```text
공통 CLI → 장치 프로필 → 명시적 경로 선택 → OS 어댑터 → AnyDesk
```

- `remote.py`: 실행 진입점
- `remote_access/`: 설정 검증, 경로 진단, OS별 앱 실행
- `examples/devices.json`: 공개 가능한 프로필 예시
- `devices.local.json`: 실제 장치 정보; Git 제외
- `.local/`: 생성된 macOS 연결 바로가기; Git 제외
- `tests/`: 입력 검증 및 OS별 실행 계획 테스트
- `academy.command`: 기존 Mac 실행 파일; 호환을 위해 유지
- `ARCHITECTURE.md`: 이전 원격 환경 진단 기록

## 프로필

최상위 `version`은 `1`, `devices`는 별칭을 키로 갖는 객체입니다. 각 장치의 `routes`에는 이름, 제공자, 전송 방식, 주소를 지정합니다. 기본 경로는 첫 번째 항목이며 `--route`로 선택합니다.

| 항목 | 의미 |
| --- | --- |
| `provider: anydesk` | 현재 지원하는 화면 제어 앱 |
| `transport: tailscale` | 직접 접속 주소의 TCP 포트 확인 후 실행 |
| `transport: vendor-relay` | AnyDesk 주소에 `/np`를 붙여 중계 연결 요청; Tailscale 검사 없음 |
| `port` | 현재 7070만 지원; 다른 포트는 검증 오류 |
| `authentication.mode: application-managed` | AnyDesk 자체 인증 사용 |

사용자 지정 앱 위치가 필요하면 최상위 `client_paths`에 현재 OS의 경로를 지정합니다. 키는 `Darwin`, `Windows`, `Linux`입니다. macOS는 `.app` 경로, Windows/Linux는 실행 파일 경로를 사용합니다. 경로와 실제 장치 정보는 로컬 파일에만 둡니다.

직접 연결 주소는 IPv4 또는 DNS 이름을 사용합니다. 현재 IPv6와 사용자 지정 포트는 지원하지 않습니다.

하나의 실행 환경에서 여러 장치를 같은 CLI로 선택할 수 있습니다. 새 장치를 추가할 때 대상별 코드 수정은 필요하지 않습니다. `${...}` 환경변수 치환은 제공하지 않으며, 로컬 JSON에는 실제 값을 작성합니다.

## 경로와 실패 처리

- `--dry-run`: 연결 계획만 생성하며 네트워크 연결·앱 실행·바로가기 파일 생성을 하지 않습니다.
- `doctor`: 선택한 경로의 사전 조건을 점검합니다. 인증과 실제 화면 연결을 보장하지 않습니다.
- `connect`: 필요한 사전 점검 후 클라이언트를 실행합니다.
- 비상 경로는 명시적으로 선택합니다. 주 경로 실패 시 자동으로 다른 경로에 접속하지 않습니다.
- 앱 미설치, 잘못된 프로필, 네트워크 실패는 구분해서 보고합니다.

AnyDesk와 OS 버전에 따라 실행 기능은 다를 수 있습니다. macOS는 `.anydeskid` 바로가기를 사용합니다. Windows/Linux는 실행 파일에 주소 인수를 전달합니다. 실제 설치된 클라이언트가 이를 지원하는지는 해당 호스트에서 확인해야 합니다.

## 검증 범위

```sh
python3 -S -m unittest discover -s tests -v
```

2026-09-05 현재 Mac에서 공통 CLI의 주 경로 TCP 진단과 비상 경로 앱 실행을 확인했습니다. 이미 인증된 AnyDesk 세션이 열리는 것을 확인한 것이며, 신규 로그인이나 재부팅 복구 시험은 아닙니다.

Windows/Linux의 명령 생성은 모의 테스트로 확인하며 실제 OS의 연결 성공을 의미하지 않습니다. PC 전원·인터넷·원격 OS 장애는 주 경로와 비상 경로의 공통 장애 요인입니다. Tailscale 중단 및 Windows 재부팅 장애 시험은 별도입니다.

## 민감 정보와 Git

실제 주소·AnyDesk ID·암호·SSH 키를 코드나 예시에 넣지 않습니다. `devices.local.json`, `*.anydeskid`, 기존 `academy-pc-access.txt`, 생성 디렉터리 `.local/`는 Git에서 제외됩니다. 암호는 프로필에 넣지 말고 AnyDesk 자체 인증 또는 암호 관리자를 사용하세요.

```sh
git status --short --ignored
git check-ignore devices.local.json academy-pc-access.txt academy.anydeskid
```

커밋할 파일을 명시적으로 추가한 뒤 `git diff --cached`로 검토하세요. `git add -f`로 로컬 접속 정보를 강제 추가하지 마세요. `.gitignore`는 이미 커밋한 정보를 삭제하지 않으며 암호화 수단도 아닙니다. Windows에서는 로컬 파일 접근 권한을 사용자 계정으로 제한하세요.

기존 Mac 바로가기와 평문 암호 파일은 자동 삭제하거나 다른 저장소로 복제하지 않습니다. 필요하면 사용자가 암호 관리자로 이전할 수 있습니다.

## 자료

- [AnyDesk macOS CLI](https://support.anydesk.com/docs/command-line-interface-for-macos)
- [AnyDesk 중계 연결 옵션](https://support.anydesk.com/session-has-ended-unexpectedly)
- [AnyDesk 무인 접속](https://support.anydesk.com/docs/unattended-access)

## 버전과 브랜치

| 버전 | 브랜치 | 내용 |
| --- | --- | --- |
| `v1.0.0` | `main` | 기존 Mac 전용 실행기의 기준 스냅샷 |
| `v2.0.0` | `feature/portable-access-v2` | 장치 프로필·OS 어댑터 기반 공통 CLI |

v1 기준 커밋은 현재 보존된 기존 파일로 만든 스냅샷입니다. v2는 v1 위에 구현했으며 `academy.command`도 유지합니다. `VERSION`은 프로젝트 버전, 프로필의 `version: 1`은 설정 형식 버전으로 서로 다릅니다.

브랜치 전환 전 수정 파일을 저장하세요. Git에서 제외한 로컬 설정과 암호 파일은 브랜치 전환으로 복원·삭제되지 않습니다.
