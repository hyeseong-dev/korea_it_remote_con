# RemoteBridge Windows GUI 기획서

## 제품 정의

RemoteBridge는 Tailscale로 연결된 Windows 11 PC의 AnyDesk 원격 접속을 한 화면에서 준비하는 실행기다. 현재 개발 버전은 `feature/windows-gui-tailscale` 브랜치의 `v3.0.0-alpha.2`다.

## 해결하려는 문제

인터넷의 두 Windows PC를 직접 연결하려고 자체 WireGuard를 사용하면 공개 허브, 키 발급, 장비별 설정 파일, 방화벽과 장애 대응이 필요하다. 이는 비기술 사용자가 쉽게 시작한다는 제품 목표와 맞지 않는다.

이번 버전은 VPN 운영 책임을 Tailscale에 맡기고 다음 사용자 흐름에 집중한다.

1. 두 PC를 같은 Tailnet에 로그인한다.
2. 원격 PC에 AnyDesk 무인 접속을 설정한다.
3. 제어 PC에서 원격 장치 이름을 저장한다.
4. RemoteBridge가 Tailscale과 원격 포트를 확인한다.
5. 준비된 경우 AnyDesk를 원격 Tailscale 주소로 실행한다.

## 책임 분리

| 구성요소 | 책임 |
| --- | --- |
| Tailscale | 장치 등록, WireGuard 기반 암호화 경로, NAT 통과, 직접 또는 릴레이 연결 |
| RemoteBridge | 설정 검증, 상태 표시, 원격 TCP 7070 검사, AnyDesk 실행 |
| AnyDesk | 무인 접속 인증, 화면·입력 전송, 실제 원격 세션 |
| 사용자 | 최초 Tailscale 로그인과 AnyDesk 무인 접속 승인 |

## 기술 구조

```text
Wails/TypeScript GUI
        │
        ▼
Go bridge.Manager
        ├─ config.json 읽기·검증
        ├─ tailscale status --json
        ├─ tailscale up/down
        ├─ 원격주소:7070 TCP 검사
        └─ AnyDesk.exe <원격주소> 실행
```

설정에는 장치 이름, 원격 Tailscale 주소와 선택적인 실행 파일 경로만 저장한다. Tailscale 인증 정보와 AnyDesk 암호는 저장하지 않는다.

## 범위

### 이번 알파에 포함

- Windows 11 amd64 GUI
- Tailscale 표준 설치 경로 자동 검색
- `tailscale status --json` 상태 판정
- 연결 및 연결 종료
- IPv4와 MagicDNS 대상 주소
- AnyDesk TCP 7070 진단과 실행
- 이전 WireGuard 알파 설정의 원격 주소 호환 읽기

### 이번 알파에서 제외

- Tailscale 자동 설치와 자동 로그인
- 인증키 발급 및 저장
- 다중 장치 목록
- AnyDesk 암호 자동 입력
- 실제 화면 연결 성공 판정
- 자체 VPN 허브와 릴레이 서버

## 검증 기준

1. 잘못된 주소와 실행 파일 경로를 거부한다.
2. Tailscale 상태 JSON을 연결됨, 꺼짐, 로그인 필요, 시작 중으로 구분한다.
3. Tailscale이 연결되지 않으면 AnyDesk를 실행하지 않는다.
4. 원격 TCP 7070이 응답하지 않으면 AnyDesk를 실행하지 않는다.
5. 앱 설정과 로그에 인증 정보가 포함되지 않는다.
6. 실제로 서로 다른 인터넷 회선의 Windows 11 두 대에서 원격 화면 연결을 확인한다.

## 다음 단계

1. Windows 11 두 대에 설치해 인터넷 구간 실연결 시험
2. 원격 PC 재부팅 후 Tailscale unattended와 AnyDesk 무인 접속 복구 확인
3. 실패 단계별 진단 메시지 개선
4. 여러 원격 PC를 저장하고 선택하는 프로필 목록 추가
