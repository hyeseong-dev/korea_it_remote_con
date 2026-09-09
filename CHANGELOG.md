# 변경 기록

## 3.0.0-beta.1 — 2026-09-09

- 여러 원격 Windows PC 프로필의 추가·전환·삭제 지원.
- 원격 PC의 Tailscale 대역 AnyDesk TCP 7070 방화벽 설정 스크립트 추가.
- 사용자 설치 안내, 출시 점검표, ZIP·SHA-256 릴리스 빌드 스크립트 추가.

## 3.0.0-alpha.2 — 2026-09-09

- `feature/windows-gui-tailscale` 브랜치에서 Windows 11 간 실사용 구성을 단순화.
- 자체 WireGuard 허브와 `.conf` 의존을 제거하고 Tailscale 관리형 네트워크로 복귀.
- Tailscale 상태 조회·연결·종료와 MagicDNS 또는 Tailscale IPv4 대상 연결 지원.
- 기존 WireGuard 알파 설정의 장치 이름, 원격 주소와 AnyDesk 경로를 읽는 호환 처리 추가.
- Tailscale 인증 정보와 AnyDesk 암호를 RemoteBridge 설정에 저장하지 않는 경계 유지.

## 3.0.0-alpha.1 — 2026-09-08

- `feature/windows-gui-go` 브랜치에서 RemoteBridge Windows GUI 알파 버전 시작.
- Go와 Wails 기반 상태 대시보드 및 로컬 연결 설정 화면 추가.
- 공식 WireGuard for Windows 터널 서비스의 조회·설치·시작·종료 흐름 추가.
- VPN 내부 IPv4의 AnyDesk TCP 7070 진단 후 AnyDesk 직접 연결 실행.
- WireGuard 개인키와 AnyDesk 암호를 앱 설정에서 다루지 않는 보안 경계 적용.
- Go 단위 테스트, 정적 검사, 프런트엔드 빌드 및 Windows 실행 파일 구동 확인.
- 기존 Python v2 CLI와 Mac 실행기를 제거하지 않고 호환 자산으로 보존.

## 2.0.0 — 2026-09-05

- 공통 Python CLI: list, connect, doctor, dry-run.
- 장치별 로컬 JSON 프로필과 OS별 AnyDesk 실행 어댑터 분리.
- Tailscale 주 경로와 AnyDesk 중계 비상 경로를 명시적으로 선택.
- 암호 처리 없이 앱 인증 사용; 실제 접속 정보는 Git 제외.
- 14개 자동 테스트 및 Mac 진단·실행 확인. Windows/Linux는 모의 테스트 범위.
- 기존 Mac 실행기 유지. RDP 및 원격 PC 설정은 변경하지 않음.

## 1.0.0 — 2026-09-05

- 보존된 Mac 전용 실행기와 공개 설정 예시의 기준 스냅샷.
- 과거 커밋의 복원이 아닌, 버전 구분을 위해 현재 생성한 초기 기록.
