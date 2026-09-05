# Remote Access v1.0.0 — Mac 전용 실행기

기존 macOS용 AnyDesk 실행 방식을 보존한 기준 버전입니다. 이 커밋은 현재 남아 있는 기존 파일을 기준으로 만든 스냅샷이며 과거 시점의 커밋을 복원한 것은 아닙니다.

## 설정 및 실행

Mac에 Tailscale과 AnyDesk가 필요합니다. 대상 장치의 AnyDesk 무인 접속은 사전에 설정되어 있어야 합니다.

1. `academy.anydeskid.example`을 `academy.anydeskid`로 복사합니다. 기존 파일은 덮어쓰지 마세요.
2. 예시 주소를 실제 Tailscale IP로 바꿉니다.
3. `./academy.command`를 실행하거나 Finder에서 더블클릭합니다.

실행기는 TCP 7070을 검사하고 Mac의 AnyDesk로 로컬 바로가기를 엽니다. 인증은 AnyDesk에서 진행합니다. 실제 접속 파일과 암호는 Git에서 제외합니다.

## 버전 구분

- `main` / `v1.0.0`: 기존 Mac 전용 실행기
- `feature/portable-access-v2` / `v2.0.0`: 장치 프로필과 OS별 어댑터를 분리한 공통 Python CLI

v2에서는 기존 실행 파일도 호환 목적으로 유지합니다. 로컬 설정 파일은 브랜치를 바꿔도 Git이 관리하지 않으므로 별도로 보관해야 합니다.
