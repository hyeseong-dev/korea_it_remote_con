# RemoteBridge Windows GUI 개발 안내

이 디렉터리는 `feature/windows-gui-tailscale` 브랜치의 `v3.0.0-beta.1` Wails/Go 데스크톱 앱입니다. 제품 사용 방법은 저장소 루트의 [README](../README.md)와 [사용자 설치 안내](../docs/USER_SETUP.md)를 우선 참고하세요.

## 동작 범위

- 로컬 Tailscale 상태를 `tailscale status --json`으로 확인하고, 필요하면 `tailscale up --timeout=10s`를 요청합니다.
- 선택한 원격 프로필의 TCP `7070` 도달 여부를 진단합니다.
- 준비된 경우 제어 PC의 AnyDesk를 실행합니다.
- Tailscale 로그인, AnyDesk 비밀번호 입력, AnyDesk 세션 성립은 외부 공식 앱의 책임입니다.

설정은 `%AppData%\RemoteBridge\config.json`에 저장됩니다. 설정 형식은 버전 3이며, 원격 PC 프로필을 1~50개 보관합니다. 인증 정보와 AnyDesk 비밀번호는 저장하지 않습니다.

## 개발 환경

- Windows 11 amd64
- Go 1.25 이상
- Node.js 및 npm
- Wails CLI v2.15.0

```powershell
cd desktop
npm --prefix frontend install
go test ./...
wails dev
```

## 릴리스 빌드

```powershell
cd desktop
wails build -clean -platform windows/amd64 -webview2 embed
```

생성 파일은 `build/bin/RemoteBridge.exe`입니다. 저장소 루트의 `scripts/Build-Release.ps1`은 테스트, 빌드, 배포 ZIP 생성과 SHA-256 계산을 함께 수행합니다.
