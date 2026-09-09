# RemoteBridge 사용자 설치 안내

## 준비물

- 제어할 Windows 11 PC 1대
- 원격으로 둘 Windows 11 PC 1대
- 두 PC에서 사용할 Tailscale 계정
- AnyDesk 무인 접속에 사용할 비밀번호

## 원격 PC 준비

1. Tailscale for Windows를 설치하고 제어 PC와 같은 Tailnet으로 로그인합니다.
2. AnyDesk를 설치합니다.
3. AnyDesk 설정에서 무인 접속과 직접 연결을 허용하고 비밀번호를 설정합니다.
4. 관리자 PowerShell에서 배포 파일에 포함된 `Configure-RemoteBridgeRemote.ps1`을 실행합니다. 이 작업은 Tailscale 대역에서만 TCP 7070을 허용합니다.
5. Tailscale 관리 화면에서 원격 PC의 기기 이름 또는 `100.x.x.x` 주소를 확인합니다.

## 제어 PC 준비

1. Tailscale과 AnyDesk를 설치하고 같은 Tailnet으로 로그인합니다.
2. `RemoteBridge.exe`를 실행합니다.
3. 화면 오른쪽 위의 `＋` 버튼을 누릅니다.
4. 원격 PC 이름과 Tailscale 주소를 입력하고 저장합니다.
5. `연결하고 AnyDesk 열기`를 누릅니다.
6. AnyDesk가 열리면 설정한 무인 접속 비밀번호로 연결합니다.

## 문제 해결

| 증상 | 확인할 내용 |
| --- | --- |
| Tailscale을 찾지 못함 | 두 PC에 Tailscale을 설치했는지 확인합니다. |
| 로그인 필요 | Tailscale 앱에서 같은 Tailnet으로 로그인합니다. |
| 원격 PC 응답 없음 | 원격 PC 전원, Tailscale 연결, AnyDesk 실행, TCP 7070 방화벽을 확인합니다. |
| AnyDesk는 열리지만 연결 불가 | AnyDesk 무인 접속 비밀번호와 직접 연결 설정을 확인합니다. |

RemoteBridge는 Tailscale 로그인 정보나 AnyDesk 비밀번호를 저장하지 않습니다.
