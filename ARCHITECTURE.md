# 원격 접속 구조 검토

> 이후 결정: RDP는 유지하고 변경하지 않는다. 현재 구현 방향은 장치 프로필과 OS별 실행 어댑터를 분리하는 공통 AnyDesk CLI다. 아래 진단은 당시의 관측 기록이며, RDP 전환 대기 상태를 현재 작업으로 해석하지 않는다. 최신 사용법은 README.md를 따른다.

## 결정 상태

2026-09-05 현장 접속 없이 실제 연결과 읽기 전용 진단을 수행했다. 기존 화면을 유지할지, 별도 RDP 작업 계정을 사용할지는 사용자 선택 대기 상태다. RDP 전환과 새 계정 생성은 아직 하지 않았다.

## 확인한 사실

- Windows 11 Pro이며 RDP가 활성화되어 있고 NLA도 켜져 있다.
- Windows 계정 암호는 확보하지 못했다. 기존 SSH 키를 통한 관리 인증은 성공했다.
- AnyDesk, OpenSSH, Tailscale 서비스는 실행 중이며 자동 시작으로 설정되어 있다. 자동 시작 설정만 확인했으며 재부팅 복구는 검증하지 않았다.
- 초기 Tailscale 시험은 DERP 중계로 약 74~78ms였다. 재시험은 직접 연결로 약 21ms였다. 지속적인 성능 보장이나 RDP 대비 벤치마크 결과는 아니다.
- Mac의 netcheck에서는 UDP 사용 가능, 목적지별 NAT 매핑 변화 없음으로 나왔다.
- AnyDesk ID에 /np를 붙인 연결이 성공했고 앱에서 AnyDesk Relay-Server가 표시되었다. Tailscale IP를 지정하는 연결과 별개의 경로를 확인했다.
- Tailscale을 실제로 중지한 장애 시험은 아직 하지 않았다.

## 현재 권고

기존 Windows 화면을 이어서 사용하는 동안에는 Tailscale + AnyDesk를 유지한다. 별도 계정의 작업 환경이 가능하면 Tailscale + RDP를 시험하여 주 연결로 채택할지 결정한다. AnyDesk ID 중계 연결은 비상 경로로 유지한다. SSH는 관리용이며 RDP의 계정 인증을 대체하지 않는다.

RDP의 성능 우위는 동일 해상도, 동일 네트워크 경로, 동일 작업 부하에서 비교해야 한다. 단일 ping 값으로 GUI 성능이나 CPU 사용량을 판단하지 않는다.

## 비상 바로가기

로컬 `academy-fallback.anydeskid`는 AnyDesk ID와 /np 옵션을 사용한다. 이 파일은 실제 접속 정보이므로 Git에서 제외한다. 기존 `academy.command`는 Tailscale 포트 검사에 의존하므로 비상 경로에서는 사용하지 않는다.

```sh
open -a AnyDesk ./academy-fallback.anydeskid
```

이 명령은 Mac용이다. Windows/Linux에서의 실행 도구는 아직 구현하지 않았다. 대상 ID와 무인 접속 프로필은 다른 OS에서도 사용할 수 있지만 각 클라이언트에서 인증 및 화면 조작을 별도 검증해야 한다.

## 전환 완료 기준

1. 현재 사용자의 작업 화면 보존 여부를 정하고, 필요 시 명시적으로 승인된 RDP 계정을 준비한다.
2. 주 연결에서 인증, 화면 표시, 키보드·마우스 입력을 검증한다.
3. Tailscale을 사용하지 않는 상태에서 AnyDesk 비상 접속을 검증한다.
4. Windows 재부팅 후 현장 승인 없이 접속 가능한지 확인한다. 재부팅은 실행 중인 작업을 확인한 후 진행한다.
5. Mac/Windows/Linux 실행 도구에서 공통 설정과 OS별 실행 부분을 분리하고, 각 OS에서 검증된 범위를 기록한다.
6. 평문 암호는 장기적으로 암호 관리자 또는 OS 자격 증명 저장소로 이전한다.
7. 실제 주소·계정·암호는 Git 추적 대상에 포함하지 않는다.

## 보존할 원칙

- 기존 관리자 암호의 임의 재설정, NLA 해제, 공인망 RDP/SSH 포트 노출을 전환의 기본 수단으로 사용하지 않는다.
- 비상 경로를 검증하기 전에 기존 정상 연결을 제거하지 않는다.
- 새 계정은 기존 사용자와 다른 프로필 및 작업 환경이라는 점을 반영한다.
- PC 전원, 인터넷, Windows 자체 장애는 두 접속 도구의 공통 장애 요인이다.

## 자료

- https://tailscale.com/docs/reference/troubleshooting/poor-performance-tailnet
- https://support.anydesk.com/session-has-ended-unexpectedly
- https://learn.microsoft.com/en-us/windows-server/remote/remote-desktop-services/remotepc/remote-desktop-allow-access
