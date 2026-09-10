# RemoteBridge v3.0.0-beta.1

RemoteBridge는 **Tailscale로 같은 Tailnet에 연결된 Windows 11 PC를 목록으로 관리하고, 준비 상태를 확인한 뒤 AnyDesk 원격 세션을 여는 Windows 데스크톱 GUI**입니다.

> 브랜치: `feature/windows-gui-tailscale`
> 상태: Windows 11 두 대의 실제 인터넷 연결 시험을 위한 베타

RemoteBridge는 VPN이나 원격 화면을 직접 중계하지 않습니다. Tailscale이 암호화된 네트워크 경로와 NAT 통과를, AnyDesk가 원격 화면·키보드·마우스와 사용자 인증을 맡습니다. 따라서 VPS, WireGuard 키 파일, 공유기 포트 전달은 필요하지 않습니다.

## 어떤 PC에 무엇을 설치하나요?

| 역할 | 의미 | 설치 대상 |
| --- | --- | --- |
| 원격 PC | 밖에서 접속할 Windows 11 PC | Tailscale, AnyDesk, 원격 PC 방화벽 설정 |
| 제어 PC | 원격 PC의 화면을 열고 조작할 Windows 11 PC | Tailscale, AnyDesk, RemoteBridge |

두 대가 서로를 제어해야 한다면 두 PC 모두에 세 프로그램을 설치하면 됩니다. 접속하는 순간의 PC가 제어 PC이고, 접속 대상이 원격 PC입니다.

```text
제어 Windows 11                                      원격 Windows 11
┌──────────────────────────────────────┐             ┌──────────────────────────┐
│ RemoteBridge                          │             │ Tailscale                │
│  ├─ Tailscale 상태 확인               │             │ AnyDesk                  │
│  ├─ 원격 TCP 7070 준비 상태 진단      │             │ 무인 접속 인증           │
│  └─ AnyDesk 실행                      │             │ TCP 7070 방화벽 규칙     │
│ Tailscale + AnyDesk                   │             └───────────▲──────────────┘
└──────────────────┬───────────────────┘                         │
                   └───── Tailscale 암호화 네트워크 ─────────────┘
                         직접 연결 또는 Tailscale 릴레이
```

## 가장 빠른 첫 연결 시험

### 1. 두 PC에 Tailscale과 AnyDesk 설치

두 Windows 11 PC에 [Tailscale for Windows](https://tailscale.com/download/windows)와 설치형 AnyDesk를 설치합니다. 두 Tailscale 앱에는 **동일한 Tailnet**으로 로그인합니다.

원격 PC에서 AnyDesk를 열고 다음을 마칩니다.

1. AnyDesk 설정에서 **무인 접속(Unattended Access)**을 켜고 강력한 비밀번호를 설정합니다.
2. AnyDesk의 직접 연결을 사용할 수 있게 설정합니다.
3. 원격 PC의 AnyDesk 주소(ID)를 기록합니다. 이 값은 Tailscale IP와 다릅니다.

### 2. 원격 PC의 방화벽 설정

원격 PC에서 **관리자 PowerShell**을 열고, 이 저장소를 내려받은 위치에서 실행합니다.

```powershell
Set-ExecutionPolicy -Scope Process Bypass
.\scripts\Configure-RemoteBridgeRemote.ps1
```

이 스크립트는 Tailscale IPv4 대역(`100.64.0.0/10`)에서 오는 TCP `7070`만 허용하는 `RemoteBridge-AnyDesk-7070` 방화벽 규칙을 만듭니다. Tailscale과 AnyDesk가 이미 설치되어 있어야 합니다.

### 3. 원격 PC의 Tailscale 주소 확인

Tailscale 관리 콘솔의 **Machines**에서 원격 PC의 아래 값 중 하나를 확인합니다.

- MagicDNS 짧은 이름 — 예: `office-pc`
- 전체 MagicDNS 이름 — 예: `office-pc.example.ts.net`
- Tailscale IPv4 — 예: `100.64.1.2`

MagicDNS 이름을 우선 권장합니다. IP가 바뀌어도 장치 이름은 유지하기 쉽습니다.

### 4. 제어 PC에서 RemoteBridge 실행 및 연결

릴리스 파일을 받았다면 압축을 풀고 `RemoteBridge.exe`를 실행합니다. 소스 저장소에서 빌드한 경우 실행 경로는 다음과 같습니다.

```powershell
.\desktop\build\bin\RemoteBridge.exe
```

처음 열리는 설정 화면에서 원격 PC 프로필을 추가합니다.

| 항목 | 입력 예 |
| --- | --- |
| 장치 이름 | `Office PC` |
| 원격 Tailscale 주소 | `office-pc` 또는 `100.64.1.2` |
| Tailscale 실행 파일 | 표준 설치 위치라면 비워 둠 |
| AnyDesk 실행 파일 | 표준 설치 위치라면 비워 둠 |

저장 후 **연결하고 AnyDesk 열기**를 누릅니다. RemoteBridge가 Tailscale 연결과 원격 PC의 TCP `7070` 응답을 확인한 다음 AnyDesk를 엽니다. 마지막으로 AnyDesk에서 원격 PC의 AnyDesk ID를 선택하거나 입력하고, 설정한 무인 접속 비밀번호로 로그인합니다.

첫 시험은 두 PC가 서로 다른 인터넷망에 있을 때 해 보세요. 예를 들어 한쪽을 휴대폰 핫스팟에 연결하면, 단순 LAN 연결이 아니라 인터넷을 통한 Tailscale 연결인지 확인할 수 있습니다.

## 앱 화면과 상태 의미

RemoteBridge에는 최대 50개의 원격 PC 프로필을 저장할 수 있습니다. 상단 장치 선택기에서 대상 PC를 바꾸고 `+`로 새 프로필을 추가합니다.

| 표시 | 확인 범위 |
| --- | --- |
| Tailscale 연결됨 | 제어 PC의 Tailscale `BackendState`가 `Running` |
| 원격 PC 응답함 | 원격 Tailscale 주소의 TCP `7070` 연결 성공 |
| AnyDesk 준비됨 | 제어 PC에서 AnyDesk 실행 파일 발견 |
| AnyDesk 실행 | AnyDesk 실행 요청을 보냄. 인증과 화면 연결 성공 자체는 AnyDesk에서 확인 |

RemoteBridge가 Tailscale 로그인을 대신하지는 않습니다. 로그인이 필요하다는 메시지가 나오면 Tailscale 공식 앱에서 로그인한 뒤 다시 시도하세요.

## 문제 해결

| 증상 | 먼저 확인할 것 |
| --- | --- |
| Tailscale을 찾지 못함 | 제어 PC에 Tailscale을 설치했는지, 표준 경로가 아니라면 `.exe` 절대 경로를 입력했는지 확인 |
| 로그인 필요 | 두 PC가 같은 Tailnet에 로그인되어 있는지 확인 |
| 원격 PC 응답 없음 | 원격 PC 전원, Tailscale 연결, AnyDesk 직접 연결 설정, TCP `7070` 방화벽 규칙 확인 |
| AnyDesk가 열리지만 세션 연결 불가 | 원격 PC의 AnyDesk ID, 무인 접속 비밀번호, AnyDesk 권한/보안 설정 확인 |
| 다른 회사 VPN과 충돌 | Windows에서 동시에 활성화된 VPN과 Tailscale 경로 정책을 확인 |

자세한 배포 절차는 [사용자 설치 안내](docs/USER_SETUP.md), 릴리스 생성 절차는 [릴리스 체크리스트](docs/RELEASE_CHECKLIST.md)를 참고하세요.

## 보안 경계

- Tailscale 로그인 정보·인증키와 AnyDesk 무인 접속 비밀번호를 RemoteBridge 설정에 저장하지 않습니다.
- 원격 주소는 IPv4 또는 안전한 호스트 이름(MagicDNS)만 허용합니다.
- AnyDesk 실행은 검증된 실행 파일 경로와 인자 배열을 사용합니다.
- TCP `7070`은 인터넷 전체가 아닌 Tailscale IPv4 대역에서만 허용합니다.
- AnyDesk ID나 비밀번호, Tailscale 인증키가 포함된 설정 파일을 Git에 커밋하지 마세요.

## 개발 및 빌드

필요 도구: Go 1.25 이상, Node.js/npm, Wails CLI v2.15.0.

```powershell
cd desktop
npm --prefix frontend install
go test ./...
wails build -clean -platform windows/amd64 -webview2 embed
```

결과 파일은 `desktop/build/bin/RemoteBridge.exe`에 생성됩니다. 배포 ZIP과 SHA-256을 만들려면 저장소 루트에서 다음을 실행합니다.

```powershell
.\scripts\Build-Release.ps1 -Version 3.0.0-beta.1
```

## 버전과 브랜치

| 버전 | 브랜치 | 역할 |
| --- | --- | --- |
| `v1.0.0` | `main` | Mac 전용 실행기 기준 버전 |
| `v2.0.0` | `feature/portable-access-v2` | Tailscale + AnyDesk 기반 Python CLI |
| `v3.0.0-alpha.1` | `feature/windows-gui-go` | 자체 WireGuard 허브를 가정한 Windows GUI 실험 |
| `v3.0.0-beta.1` | `feature/windows-gui-tailscale` | Tailscale + AnyDesk 기반 Windows GUI |

## 베타 제한 사항

- Windows 11 amd64를 실제 배포 대상으로 합니다.
- Tailscale·AnyDesk의 설치와 최초 로그인, AnyDesk 무인 접속 설정은 사용자가 수행합니다.
- TCP `7070` 응답은 직접 연결 경로의 준비 상태일 뿐, AnyDesk 인증이나 화면 연결 완료를 보장하지는 않습니다.
- 실제 인터넷망의 Windows 11 두 대에서 장시간 연결·재부팅 복구 검증이 남아 있습니다.
