# RemoteBridge v3.0.0-alpha.1 WireGuard 실험 아키텍처

> 이 문서는 `feature/windows-gui-go` 브랜치에서 진행한 이전 WireGuard 허브 실험을 보존한다. 현재 `feature/windows-gui-tailscale` 브랜치의 `v3.0.0-alpha.2`는 자체 허브 대신 Tailscale을 사용한다. 현재 설계는 [상위 기획서](../REMOTE_BRIDGE_PLAN.md)를 참고한다.

![RemoteBridge v3.0.0-alpha.1 현재 구현 아키텍처](./remote-bridge-v3-alpha1.png)

## 핵심 역할

RemoteBridge는 제어 Windows 11 PC에서 실행되는 연결 준비 GUI다. VPN 프로토콜이나 화면 전송 기능을 직접 구현하지 않고, 공식 WireGuard for Windows와 AnyDesk를 순서대로 제어한다.

1. `%AppData%\RemoteBridge\config.json`에서 원격 VPN 주소, 터널 이름과 WireGuard 설정 파일 경로를 읽고 검증한다.
2. `WireGuardTunnel$<터널 이름>` Windows 서비스 상태를 확인한다.
3. 서비스가 없으면 `wireguard.exe /installtunnelservice <설정 파일>`로 설치하고, 중지 상태면 시작한다.
4. VPN 내부 원격 주소의 TCP 7070 응답을 최대 3초 동안 확인한다.
5. 응답하는 경우에만 `AnyDesk.exe <원격 VPN 주소>`를 실행한다.
6. 사용자 인증과 화면·키보드·마우스 전송은 AnyDesk가 담당한다.

## 외부에서 준비해야 하는 요소

현재 알파 버전은 다음 항목을 자동으로 만들지 않는다.

- 인터넷에서 접근 가능한 WireGuard 허브
- 장비별 WireGuard 키와 `.conf` 또는 `.conf.dpapi`
- VPN 주소 배정과 허브 Peer 등록
- 원격 Windows의 AnyDesk TCP 7070 방화벽 규칙
- AnyDesk 무인 접속 인증

따라서 현재 구현은 전체 VPN 관리 시스템이 아니라, 준비된 WireGuard 네트워크 위에서 연결 상태를 진단하고 AnyDesk를 안전한 순서로 실행하는 Windows GUI다.
