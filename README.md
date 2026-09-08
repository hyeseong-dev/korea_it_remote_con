# RemoteBridge v3.0.0-alpha.1

RemoteBridge는 **공식 WireGuard VPN 터널을 준비하고, VPN 내부 주소의 Windows 11 PC에 AnyDesk 직접 연결을 여는 데스크톱 GUI**입니다.

이 브랜치의 목표는 Tailscale에 의존하던 기존 실행 흐름을 자체 운영 가능한 WireGuard 네트워크로 옮기는 것입니다. VPN 암호화 프로토콜 자체를 새로 만들지는 않습니다. 검증된 WireGuard를 데이터 전송 계층으로 사용하고, 이 프로젝트는 설정·상태 진단·AnyDesk 실행 경험을 제공합니다.

> 현재 상태: Windows 11용 기능 검증 단계입니다. 실제 WireGuard 허브 구성, 키 발급 및 배포는 아직 자동화하지 않습니다.

## 프로그램 기획

기존 v2는 Tailscale과 AnyDesk가 이미 설치되고 연결된 환경에서 Python 명령을 실행해야 했습니다. v3는 비기술 사용자도 한 화면에서 다음 상태를 확인하고 접속할 수 있도록 설계합니다.

1. 공식 WireGuard 터널이 설치되어 있는지 확인합니다.
2. 중지된 터널을 시작하거나, 로컬 설정 파일로 최초 터널 서비스를 설치합니다.
3. VPN 내부 IPv4에서 AnyDesk 기본 포트 `7070`이 응답하는지 확인합니다.
4. 응답이 확인된 경우에만 AnyDesk를 해당 VPN 주소로 실행합니다.
5. VPN, 원격 Windows PC, AnyDesk 준비 상태를 GUI에 구분해 표시합니다.

상세한 문제 정의, 범위와 단계별 개발 계획은 [RemoteBridge 제품 기획서](docs/REMOTE_BRIDGE_PLAN.md)를 참고하세요.

## 목적

- Tailscale 계정과 제어 서버에 대한 필수 의존성을 줄입니다.
- 원격 접속 전에 VPN과 대상 포트를 진단하여 실패 원인을 구분합니다.
- WireGuard와 AnyDesk의 복잡한 실행 절차를 하나의 Windows GUI로 묶습니다.
- 기존 AnyDesk 인증과 화면 전송 기능을 재사용합니다.
- 개인키, AnyDesk 암호와 실제 장치 정보가 Git에 들어가지 않도록 분리합니다.

## 동작 구조

```text
RemoteBridge GUI (Windows 11)
        │
        ├─ 공식 WireGuard 터널 서비스 시작/조회
        │
        ▼
WireGuard 사설망 ─────────────── 원격 Windows 11
                                  │
                                  └─ AnyDesk TCP 7070
        │
        └─ 포트 응답 확인 후 AnyDesk.exe <VPN-IP> 실행
```

초기 배포는 공인 IP가 있는 별도 WireGuard 허브를 두고 각 장치가 허브로 접속하는 방식을 전제로 합니다. NAT 통과, P2P 연결 조정, 릴레이 서버는 이번 알파 범위에 포함하지 않습니다.

## 현재 제공 기능

- Windows WireGuard 터널 서비스 상태 확인
- 기존 터널 서비스 시작 및 종료
- `.conf` 또는 `.conf.dpapi`를 통한 최초 터널 서비스 설치
- 원격 VPN IPv4의 TCP 7070 연결 진단
- 진단 성공 후 AnyDesk 직접 연결 실행
- WireGuard 및 AnyDesk 표준 설치 경로 자동 검색
- 로컬 설정 저장과 입력값 검증
- 한국어 상태 대시보드와 연결 설정 화면

## 지원 범위

| 구분 | 현재 상태 |
| --- | --- |
| GUI 실행 환경 | Windows 11 amd64 |
| VPN 엔진 | 공식 WireGuard for Windows |
| 원격 대상 | WireGuard와 AnyDesk가 준비된 Windows 11 PC |
| 화면 제어 | AnyDesk 직접 연결, TCP 7070 |
| macOS/Linux GUI | 계획됨, 미구현 |
| VPN 허브·키 자동 발급 | 계획됨, 미구현 |
| 자체 NAT 통과·릴레이 | 장기 검토 |
| AnyDesk 인증 자동화 | 의도적으로 제외 |

## 빠른 실행

### 1. 사전 준비

로컬 Windows 11 PC에 다음 프로그램을 설치합니다.

- 공식 WireGuard for Windows
- 설치형 AnyDesk
- WebView2 Runtime — Windows 11에는 일반적으로 포함됨

원격 Windows 11 PC에서는 WireGuard 터널과 AnyDesk 무인 접속을 별도로 설정하고, AnyDesk의 직접 연결 허용 및 TCP 7070 방화벽 규칙을 준비해야 합니다.

### 2. 프로그램 실행

빌드된 파일을 실행합니다.

```powershell
.\desktop\build\bin\RemoteBridge.exe
```

첫 실행 시 자동으로 열리는 설정 화면에 다음 값을 입력합니다.

| 항목 | 설명 | 예시 |
| --- | --- | --- |
| 장치 이름 | 화면에 표시할 원격 PC 이름 | `Academy PC` |
| VPN 내부 IPv4 | WireGuard가 원격 PC에 할당한 주소 | `10.88.0.2` |
| 터널 이름 | WireGuard 터널 서비스 이름 | `academy` |
| WireGuard 설정 파일 | 로컬 `.conf` 또는 `.conf.dpapi` 절대 경로 | `C:\ProgramData\RemoteBridge\academy.conf.dpapi` |
| WireGuard 실행 파일 | 비워 두면 표준 경로에서 검색 | 선택 사항 |
| AnyDesk 실행 파일 | 비워 두면 표준 경로에서 검색 | 선택 사항 |

설정을 저장한 뒤 **연결하고 AnyDesk 열기**를 누릅니다. 최초 터널 설치와 서비스 시작·종료에는 Windows 관리자 권한이 필요할 수 있습니다.

실제 설정은 `%AppData%\RemoteBridge\config.json`에 저장됩니다. 이 파일에는 WireGuard 개인키나 AnyDesk 암호가 아니라 실행 경로와 연결 대상만 들어갑니다.

## 개발 및 빌드

요구 사항:

- Go 1.25 이상
- Node.js와 npm
- Wails CLI v2.15
- Windows용 WebView2 개발 환경

```powershell
git clone https://github.com/hyeseong-dev/korea_it_remote_con.git
cd korea_it_remote_con
git fetch origin feature/windows-gui-go
git switch --track origin/feature/windows-gui-go

go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
cd desktop
npm --prefix frontend install
go test ./...
go vet ./...
wails build -clean -platform windows/amd64 -webview2 embed
```

빌드 결과는 `desktop/build/bin/RemoteBridge.exe`입니다. 자세한 개발 방법과 설정 위치는 [desktop/README.md](desktop/README.md)를 참고하세요.

## 보안 원칙

- WireGuard 개인키, 실제 `.conf`, `.conf.dpapi`, AnyDesk 암호를 커밋하지 않습니다.
- 앱은 VPN 암호화나 AnyDesk 인증을 직접 구현하지 않습니다.
- 외부 프로그램은 셸 문자열이 아니라 검증된 실행 파일과 인자 배열로 실행합니다.
- AnyDesk는 VPN 내부 주소의 TCP 7070 응답이 확인된 경우에만 실행합니다.
- 원격 Windows 방화벽은 가능하면 WireGuard 인터페이스와 VPN 대역에서만 7070을 허용합니다.

## 저장소와 브랜치

저장소: `hyeseong-dev/korea_it_remote_con`

| 버전 | 브랜치 | 역할 |
| --- | --- | --- |
| `v1.0.0` | `main` | 보존된 Mac 전용 실행기 기준 버전 |
| `v2.0.0` | `feature/portable-access-v2` | Tailscale + AnyDesk 기반 Python 공통 CLI |
| `v3.0.0-alpha.1` | `feature/windows-gui-go` | WireGuard + AnyDesk 기반 Windows Go GUI |

`feature/windows-gui-go`는 GitHub 원격 브랜치로 공개되어 있으며, 위 명령으로 다른 PC에서도 동일한 개발 버전을 체크아웃할 수 있습니다.

기존 Python CLI와 실행기는 이 브랜치에서도 제거하지 않았습니다. v2의 기준 구현과 설명은 `feature/portable-access-v2` 브랜치에서 확인할 수 있습니다.

## 디렉터리

```text
desktop/                    Go/Wails Windows GUI
  frontend/                 TypeScript 화면
  internal/bridge/          설정, 진단, WireGuard·AnyDesk 제어
  config.example.json       공개 가능한 설정 형식 예시
remote_access/              기존 Python v2 코어
tests/                      기존 Python v2 테스트
docs/REMOTE_BRIDGE_PLAN.md  v3 제품 기획 및 로드맵
```

## 알려진 제한

- 실제 VPS WireGuard 허브 및 운영 키가 없으면 원격 연결은 완료되지 않습니다.
- 터널 설치·서비스 제어는 Windows 권한 정책에 따라 관리자 실행이 필요합니다.
- `launched` 또는 GUI의 실행 완료 문구는 AnyDesk 인증과 화면 연결 성공을 보장하지 않습니다.
- Windows 재부팅 복구, 키 교체, 허브 장애 전환과 설치 프로그램 배포는 아직 검증하지 않았습니다.
