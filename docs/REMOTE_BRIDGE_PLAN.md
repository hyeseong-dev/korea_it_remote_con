# RemoteBridge 제품 기획서

## 1. 제품 개요

RemoteBridge는 WireGuard 기반 사설망과 AnyDesk를 하나의 연결 경험으로 묶는 원격 접속 런처입니다. 사용자는 VPN 클라이언트와 원격 제어 프로그램의 세부 동작을 각각 다루는 대신, RemoteBridge에서 장치 상태를 확인하고 한 번의 동작으로 연결을 시작합니다.

현재 개발 버전은 `feature/windows-gui-go` 브랜치의 `v3.0.0-alpha.1`입니다. Windows 11 제어 PC와 Windows 11 원격 PC를 우선 대상으로 합니다.

## 2. 해결하려는 문제

기존 환경은 Tailscale, AnyDesk, 운영체제 설정을 각각 이해해야 했습니다. 연결이 실패하면 VPN 장애인지, 원격 PC가 꺼진 것인지, AnyDesk가 준비되지 않은 것인지 빠르게 구분하기도 어려웠습니다.

RemoteBridge는 다음 문제를 해결합니다.

- 원격 접속 전에 네트워크와 AnyDesk 포트를 단계적으로 진단합니다.
- WireGuard 터널의 시작과 AnyDesk 실행을 하나의 흐름으로 연결합니다.
- 상태를 기술 용어만으로 표시하지 않고 사용자가 다음 행동을 판단할 수 있는 문장으로 제공합니다.
- 실제 접속 정보와 인증정보를 소스 코드에서 분리합니다.
- Tailscale 제어 서버 없이 운영 가능한 단순한 VPN 구성을 지향합니다.

## 3. 목표 사용자와 사용 시나리오

### 목표 사용자

- 학원, 사무실 또는 집의 Windows 11 PC에 반복적으로 접속하는 사용자
- WireGuard와 AnyDesk 설치는 관리할 수 있지만 매번 명령을 실행하고 싶지 않은 사용자
- VPN 키와 원격 제어 암호를 Git이나 자체 앱 서버에 저장하고 싶지 않은 소규모 운영자

### 대표 시나리오

1. 사용자가 RemoteBridge를 실행합니다.
2. 앱이 WireGuard 터널, 원격 장치, AnyDesk 상태를 표시합니다.
3. 사용자가 **연결하고 AnyDesk 열기**를 누릅니다.
4. 앱이 WireGuard 터널을 시작합니다.
5. 앱이 원격 VPN 주소의 TCP 7070 응답을 확인합니다.
6. 응답이 있으면 AnyDesk를 VPN 내부 주소로 실행합니다.
7. 실제 사용자 인증과 화면 제어는 AnyDesk에서 완료합니다.

## 4. 제품 원칙

### 검증된 암호화 사용

새 VPN 암호화 프로토콜을 만들지 않습니다. 공식 WireGuard 구현을 사용하고 RemoteBridge는 그 위의 제어 계층을 담당합니다.

### 인증 책임 분리

WireGuard 개인키는 WireGuard 설정에, 원격 화면 인증은 AnyDesk에 맡깁니다. RemoteBridge 설정에는 개인키와 암호를 넣지 않습니다.

### 자동 우회 금지

VPN 직접 연결 실패 시 AnyDesk 공급자 중계로 몰래 전환하지 않습니다. 후속 버전에서 비상 경로를 제공하더라도 사용자가 명시적으로 선택하도록 합니다.

### 실패 원인 구분

다음 상태를 별도로 표시합니다.

- 설정되지 않음
- WireGuard 미설치 또는 터널 미등록
- VPN 중지 또는 시작 실패
- VPN 연결됨, AnyDesk 포트 응답 없음
- AnyDesk 실행 파일 없음
- AnyDesk 실행 요청 완료

## 5. 기술 구조

```text
┌──────────────────────────────────────────┐
│ RemoteBridge — Go + Wails                │
│                                          │
│  TypeScript UI                           │
│    └─ 상태 카드 / 설정 / 연결 버튼       │
│              │                           │
│  Go application boundary                │
│              │                           │
│  bridge.Manager                          │
│    ├─ 엄격한 JSON 설정 검증              │
│    ├─ Windows 서비스 상태 조회           │
│    ├─ WireGuard 터널 시작·종료           │
│    ├─ TCP 7070 제한시간 진단             │
│    └─ AnyDesk 프로세스 실행              │
└──────────────────────────────────────────┘
              │
              ▼
   공식 WireGuard for Windows
              │
              ▼
       WireGuard VPN 허브
              │
              ▼
 원격 Windows 11 — AnyDesk TCP 7070
```

### 기술 선택

| 영역 | 선택 | 이유 |
| --- | --- | --- |
| 백엔드 | Go | 단일 실행 파일, 네트워크·서비스 제어, 명시적 오류 처리 |
| GUI | Wails + Vanilla TypeScript | Windows WebView2 활용, 작은 프런트엔드 의존성 |
| VPN | 공식 WireGuard | 검증된 암호화와 Windows 터널 서비스 제공 |
| 원격 화면 | AnyDesk | 기존 무인 접속과 화면 제어 기능 유지 |
| 설정 | 로컬 JSON | 배포 초기 단계에서 구조가 단순하고 검토 가능 |

## 6. 네트워크 기획

### 알파 권장 구성: 허브 앤 스포크

공인 IP가 있는 VPS를 WireGuard 허브로 두고 제어 PC와 원격 Windows PC가 모두 허브로 아웃바운드 연결합니다.

예시 주소 계획:

| 역할 | VPN 주소 예시 |
| --- | --- |
| WireGuard 허브 | `10.88.0.1` |
| 원격 Windows 11 | `10.88.0.2` |
| 제어 Windows 11 | `10.88.0.3` |

원격 Windows 방화벽에서는 TCP 7070을 전체 인터넷이 아니라 VPN 대역 또는 WireGuard 인터페이스에만 허용합니다.

### 후속 검토

- 장치 등록과 공개키 승인 API
- 키 만료와 교체
- 허브 이중화
- 직접 P2P 경로 탐색
- 직접 연결 실패 시 암호화된 릴레이
- macOS와 Linux 클라이언트

직접 P2P, NAT 통과와 릴레이를 구현하면 Tailscale과 유사한 제어면 규모가 되므로 MVP와 분리합니다.

## 7. 설정과 데이터

기본 설정 위치는 `%AppData%\RemoteBridge\config.json`입니다.

저장하는 값:

- 표시용 장치 이름
- 원격 VPN IPv4
- WireGuard 터널 이름
- 로컬 WireGuard 설정 파일 경로
- 선택적인 WireGuard 및 AnyDesk 실행 파일 경로

저장하지 않는 값:

- WireGuard 개인키 내용
- AnyDesk 무인 접속 암호
- Windows 계정 암호
- VPS 관리자 자격증명

WireGuard 설정 파일 자체에는 개인키가 들어갈 수 있으므로 저장소 밖에서 관리하고, 가능하면 WireGuard for Windows의 DPAPI 암호화 형식을 사용합니다.

## 8. MVP 범위와 완료 기준

### 구현 완료

- Windows GUI 골격
- 로컬 설정 입력 및 검증
- WireGuard 터널 서비스 조회·설치·시작·종료
- TCP 7070 진단
- AnyDesk 직접 실행
- 단위 테스트와 Windows amd64 빌드

### 운영 검증 필요

- 실제 VPS 허브와 두 피어의 통신
- VPN 주소를 통한 AnyDesk 직접 연결 확인
- Windows 재부팅 후 터널과 AnyDesk 준비 상태
- 관리자 권한 요청 경험
- WireGuard 미설치와 설정 오류 안내
- 원격 PC 오프라인 상태 안내

알파 완료는 위 운영 검증을 통과하고, 개인키가 로그·설정·Git 기록에 노출되지 않았음을 확인했을 때로 정의합니다.

## 9. 로드맵

### 3.0.0-alpha.1 — 로컬 GUI MVP

- Windows GUI와 연결 오케스트레이션
- 수동으로 준비된 WireGuard 설정 사용

### 3.0.0-alpha.2 — 실제 VPN 통합 시험

- VPS 허브 구축 문서
- Windows 원격 호스트 방화벽 정책
- 실제 연결·재연결·재부팅 시험
- 오류 코드와 진단 로그 개선

### 3.0.0-beta.1 — 배포 준비

- Windows 설치 프로그램
- 관리자 권한 도우미 분리
- 설정 백업과 마이그레이션
- 코드 서명 및 업데이트 전략

### 3.x 후속

- macOS/Linux GUI 또는 공통 CLI
- 장치 다중 프로필
- 키 등록·회수 제어면
- 허브 장애 전환

## 10. 비목표

현재 버전에서 다음 기능은 만들지 않습니다.

- 새로운 암호화 알고리즘이나 VPN 프로토콜
- AnyDesk 화면 프로토콜 대체
- AnyDesk 암호 자동 입력
- Windows 계정 생성·암호 변경
- 공인 인터넷에 AnyDesk 포트 직접 노출
- 사용자 모르게 공급자 중계 경로로 전환

## 11. 저장소 운영

- 기준 저장소: `https://github.com/hyeseong-dev/korea_it_remote_con.git`
- 개발 브랜치: `feature/windows-gui-go`
- v2 기준 브랜치: `feature/portable-access-v2`
- v1 기준 브랜치: `main`

기능 개발은 `feature/windows-gui-go`에서 진행합니다. 실제 접속 주소, `.conf`, `.conf.dpapi`, 키와 암호는 커밋하지 않습니다. 릴리스 준비 시 알파 태그와 변경 기록을 검토한 뒤 별도의 통합 또는 릴리스 전략을 결정합니다.
