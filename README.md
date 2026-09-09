# RemoteBridge v3.0.0-beta.1

RemoteBridge는 **Tailscale로 연결된 Windows 11 PC의 상태를 확인하고 AnyDesk 원격 화면을 여는 데스크톱 GUI**다.

이 버전은 자체 VPN 허브를 운영하지 않는다. Tailscale이 장치 등록, 암호화된 네트워크 경로, NAT 통과와 필요 시 릴레이를 담당하고, RemoteBridge는 비기술 사용자가 연결 상태를 확인하고 AnyDesk를 실행하는 한 화면을 제공한다.

> 현재 브랜치: `feature/windows-gui-tailscale`
>
> 현재 상태: Windows 11 간 배포 검증용 베타 버전

## 목적

- VPS, WireGuard 키와 `.conf`, 공유기 포트 전달 없이 인터넷의 두 Windows PC를 연결한다.
- VPN 상태, 원격 PC 도달 여부와 AnyDesk 준비 상태를 분리해 표시한다.
- 원격 장치 주소를 확인한 뒤 한 번의 동작으로 AnyDesk를 실행한다.
- Tailscale 인증 정보와 AnyDesk 암호는 각각의 공식 앱에 맡긴다.

## 아키텍처

```text
제어 Windows 11                              원격 Windows 11
┌─────────────────────────┐                ┌────────────────────┐
│ RemoteBridge             │                │ Tailscale          │
│  ├─ Tailscale 상태 확인  │                │ AnyDesk TCP 7070   │
│  ├─ 원격 7070 진단       │                │ 무인 접속 인증     │
│  └─ AnyDesk 실행         │                └────────────────────┘
│ Tailscale + AnyDesk      │                         ▲
└────────────┬────────────┘                         │
             └──── Tailscale 암호화 네트워크 ──────┘
                   직접 연결 또는 관리형 릴레이
```

RemoteBridge가 화면 데이터를 중계하지는 않는다. Tailscale이 네트워크 경로를 만들고, AnyDesk가 사용자 인증과 화면·키보드·마우스 전송을 담당한다.

## 빠른 실행

### 1. 두 PC 준비

제어 PC와 원격 PC에 다음 프로그램을 설치한다.

- [Tailscale for Windows](https://tailscale.com/download/windows)
- 설치형 AnyDesk

두 PC를 같은 Tailnet에 로그인한다. 원격 PC에는 AnyDesk 무인 접속을 설정하고 직접 연결을 허용한다. 필요한 경우 Windows 방화벽에서 Tailscale 네트워크의 TCP 7070을 허용한다.

### 2. 원격 주소 확인

Tailscale 관리 콘솔의 Machines 화면에서 원격 PC의 다음 값 중 하나를 확인한다.

- Tailscale IPv4: 예) `100.64.1.2`
- MagicDNS 이름: 예) `academy-pc`
- 전체 MagicDNS 이름: 예) `academy-pc.example.ts.net`

### 3. 프로그램 실행

```powershell
.\desktop\build\bin\RemoteBridge.exe
```

첫 설정 화면에 다음 값을 입력한다.

| 항목 | 예시 |
| --- | --- |
| 장치 이름 | `Academy PC` |
| 원격 Tailscale 주소 | `academy-pc` 또는 `100.64.1.2` |
| Tailscale 실행 파일 | 표준 설치라면 비움 |
| AnyDesk 실행 파일 | 표준 설치라면 비움 |

설정을 저장하고 **연결하고 AnyDesk 열기**를 누른다. Tailscale 로그인이 필요하다는 메시지가 나오면 공식 Tailscale 앱에서 로그인한 후 다시 시도한다.

자세한 실행 및 빌드 방법은 [desktop/README.md](desktop/README.md)를 참고한다.

## 상태 판정

| 표시 | 의미 |
| --- | --- |
| Tailscale 연결됨 | 로컬 Tailscale `BackendState`가 `Running` |
| 원격 PC 응답함 | 원격 주소의 TCP 7070 연결 성공 |
| AnyDesk 준비됨 | 로컬 AnyDesk 실행 파일 발견 |
| AnyDesk 실행 | 실행 요청 전달 완료. 인증·화면 연결 성공과는 구분 |

## 개발과 빌드

```powershell
cd desktop
npm --prefix frontend install
go test ./...
wails build -clean -platform windows/amd64 -webview2 embed
```

빌드 결과:

```text
desktop/build/bin/RemoteBridge.exe
```

## 보안 경계

- Tailscale 인증키나 로그인 정보를 앱 설정에 저장하지 않는다.
- AnyDesk 암호를 앱 설정이나 실행 인수에 넣지 않는다.
- 원격 주소는 IPv4 또는 안전한 호스트 이름 형식만 허용한다.
- 외부 실행은 셸 문자열이 아닌 검증된 실행 파일과 인자 배열을 사용한다.
- TCP 7070은 가능하면 Tailscale 네트워크에서만 허용한다.

## 버전과 브랜치

| 버전 | 브랜치 | 역할 |
| --- | --- | --- |
| `v1.0.0` | `main` | Mac 전용 실행기 기준 버전 |
| `v2.0.0` | `feature/portable-access-v2` | Tailscale + AnyDesk 기반 Python CLI |
| `v3.0.0-alpha.1` | `feature/windows-gui-go` | 자체 WireGuard 허브를 가정한 Windows GUI 실험 |
| `v3.0.0-beta.1` | `feature/windows-gui-tailscale` | Tailscale + AnyDesk 기반 Windows GUI |

## 현재 제한

- Windows 11 amd64만 실제 배포 대상으로 한다.
- 원격 PC 프로필을 최대 50개까지 저장한다.
- Tailscale과 AnyDesk의 설치 및 최초 로그인은 자동화하지 않는다.
- 원격 PC가 꺼져 있거나 AnyDesk 7070이 차단되면 실행하지 않는다.
- 현재 Windows 11 두 대의 실제 인터넷 연결 검증은 남아 있다.
