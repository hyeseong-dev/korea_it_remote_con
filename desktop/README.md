# RemoteBridge desktop MVP

Windows 11에서 공식 WireGuard 터널을 시작한 뒤 VPN 내부 주소의 AnyDesk TCP 7070 연결을 확인하고 AnyDesk를 실행하는 Wails 기반 GUI입니다.

이 디렉터리는 저장소의 `feature/windows-gui-go` 브랜치에서 개발하는 `v3.0.0-alpha.1` 구현입니다. 루트의 Python 파일은 `feature/portable-access-v2`에서 시작된 기존 v2 구현이며, 마이그레이션 비교와 호환을 위해 남겨 둡니다.

## 보안 경계

- VPN 프로토콜이나 암호화를 직접 구현하지 않습니다.
- WireGuard 개인키와 AnyDesk 암호를 앱 설정에 저장하지 않습니다.
- WireGuard 설정 파일은 로컬에만 두고 Git에 커밋하지 않습니다.
- 외부 명령은 셸 없이 고정된 인자 배열로 실행합니다.
- AnyDesk 실행 전 TCP 7070 응답을 확인합니다.

## 사전 조건

- Windows 11
- 공식 WireGuard for Windows
- 설치형 AnyDesk와 직접 연결 허용
- 미리 준비한 WireGuard `.conf` 또는 `.conf.dpapi` 파일
- 터널 설치와 시작/종료를 위한 관리자 권한

## 개발

프로젝트 루트의 `.local` 도구는 Git에서 제외됩니다. 시스템에 Go와 Wails가 설치돼 있다면 일반 명령을 사용해도 됩니다.

```powershell
cd desktop
npm --prefix frontend install
go test ./...
wails dev
wails build -clean -platform windows/amd64 -webview2 embed
```

빌드 결과는 `desktop/build/bin/RemoteBridge.exe`에 생성됩니다. 첫 실행 시 설정 창에서 장치 이름, VPN 내부 IPv4, WireGuard 터널 이름과 설정 파일 경로를 입력합니다. 실행 파일 경로를 비워 두면 표준 설치 위치에서 자동으로 찾습니다.

## 사용자 구동 순서

1. 공식 WireGuard for Windows와 설치형 AnyDesk를 설치합니다.
2. 저장소 밖에 WireGuard `.conf` 또는 `.conf.dpapi` 파일을 준비합니다.
3. `RemoteBridge.exe`를 실행합니다.
4. 장치 이름, 원격 VPN IPv4, 터널 이름과 설정 파일 절대 경로를 저장합니다.
5. **연결하고 AnyDesk 열기**를 누릅니다.
6. AnyDesk가 열리면 앱 자체의 무인 접속 인증을 완료합니다.

최초 터널 설치와 서비스 시작·종료는 Windows 관리자 권한이 필요할 수 있습니다. RemoteBridge는 권한 상승을 우회하지 않습니다.

실제 설정은 `%AppData%\RemoteBridge\config.json`에 저장됩니다. 이 파일에는 개인키가 아니라 로컬 WireGuard 설정 파일의 경로만 들어갑니다. 테스트나 별도 프로필이 필요하면 `REMOTE_BRIDGE_CONFIG` 환경 변수로 설정 파일 경로를 바꿀 수 있습니다.

## 현재 범위

- Windows WireGuard 터널 서비스 상태 확인
- 터널 서비스 시작 및 종료
- 최초 연결 시 공식 WireGuard 실행 파일을 통한 터널 서비스 설치
- VPN 내부 IPv4의 AnyDesk TCP 7070 진단
- 진단 성공 후 AnyDesk 직접 연결 실행

키 발급, VPS 허브 프로비저닝, NAT 통과, 릴레이, 자동 업데이트는 후속 단계입니다.
