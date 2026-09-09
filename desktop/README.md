# RemoteBridge Windows GUI

`feature/windows-gui-tailscale` 브랜치의 `v3.0.0-alpha.2` 구현이다. Windows 11 제어 PC에서 Tailscale 연결 상태를 확인하고, 원격 Windows 11의 AnyDesk TCP 7070이 응답할 때 AnyDesk를 실행한다.

## 준비 사항

두 Windows 11 PC에 다음 프로그램을 설치한다.

- [Tailscale for Windows](https://tailscale.com/download/windows)
- 설치형 AnyDesk

두 PC를 같은 Tailnet에 로그인한다. 원격 PC에서는 AnyDesk의 직접 연결과 무인 접속을 활성화하고, Windows 방화벽에서 TCP 7070을 Tailscale 네트워크에 허용한다.

## 사용자 실행 순서

1. 제어 PC와 원격 PC의 Tailscale 상태가 연결됨인지 확인한다.
2. 제어 PC에서 `desktop/build/bin/RemoteBridge.exe`를 실행한다.
3. 장치 이름과 원격 PC의 Tailscale IPv4 또는 MagicDNS 이름을 입력한다.
4. 실행 파일 경로는 표준 위치에 설치했다면 비워 둔다.
5. **연결하고 AnyDesk 열기**를 누른다.
6. AnyDesk에서 무인 접속 인증을 완료한다.

MagicDNS가 활성화된 Tailnet에서는 IP 대신 `academy-pc` 또는 전체 이름인 `academy-pc.example.ts.net`을 사용할 수 있다. 주소는 Tailscale 관리 콘솔의 Machines 화면이나 `tailscale status`에서 확인한다.

## 앱 동작 순서

1. `tailscale status --json`으로 로컬 연결 상태를 확인한다.
2. 꺼져 있으면 `tailscale up --timeout=10s`로 연결한다.
3. 원격 주소의 TCP 7070을 최대 3초 동안 확인한다.
4. 응답하는 경우에만 `AnyDesk.exe <원격 주소>`를 실행한다.
5. **VPN 종료**을 누르면 `tailscale down`을 실행한다.

로그인이 필요하면 RemoteBridge가 인증을 대신하지 않는다. Tailscale 앱에서 로그인한 뒤 다시 시도한다. Tailscale 로그인 정보와 AnyDesk 암호는 RemoteBridge에 저장하지 않는다.

## 설정 파일

설정은 `%AppData%\RemoteBridge\config.json`에 저장된다.

```json
{
  "version": 2,
  "deviceLabel": "Academy PC",
  "targetAddress": "academy-pc.example.ts.net",
  "anyDeskPort": 7070,
  "tailscalePath": "C:\\Program Files\\Tailscale\\tailscale.exe",
  "anyDeskPath": "C:\\Program Files (x86)\\AnyDesk\\AnyDesk.exe"
}
```

`v3.0.0-alpha.1`의 WireGuard 설정이 남아 있으면 장치 이름, 원격 주소와 AnyDesk 경로를 메모리에서 변환해 읽는다. 다음 저장부터 새 형식으로 기록된다.

## 개발과 빌드

요구 사항:

- Go 1.25 이상
- Node.js와 npm
- Wails CLI v2.15.0

```powershell
cd desktop
npm --prefix frontend install
go test ./...
wails dev
wails build -clean -platform windows/amd64 -webview2 embed
```

빌드 결과는 `desktop/build/bin/RemoteBridge.exe`에 생성된다.

## 현재 제한

- Windows 11 amd64만 실제 배포 대상으로 한다.
- 한 번에 원격 장치 프로필 하나만 저장한다.
- AnyDesk 기본 직접 연결 포트 7070만 지원한다.
- Tailscale 설치와 최초 로그인, AnyDesk 무인 접속 설정은 사용자가 수행한다.
- TCP 응답과 앱 실행은 확인하지만 AnyDesk 인증 및 화면 연결 완료까지 판정하지 않는다.
