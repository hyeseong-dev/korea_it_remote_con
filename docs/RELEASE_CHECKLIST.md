# 출시 점검표

- [ ] Windows 11 제어 PC와 원격 PC가 서로 다른 인터넷 회선에서 연결된다.
- [ ] 원격 PC 재부팅 후 Tailscale과 AnyDesk 무인 접속이 복구된다.
- [ ] Tailscale 미설치, 로그인 필요, 원격 PC 꺼짐, 방화벽 차단을 각각 안내한다.
- [ ] 원격 PC 설정 스크립트가 TCP 7070을 `100.64.0.0/10`으로만 제한한다.
- [ ] `scripts/Build-Release.ps1`로 ZIP과 SHA-256을 생성한다.
- [ ] Windows Defender SmartScreen과 백신 오탐을 확인한다.
- [ ] 코드 서명 인증서로 EXE와 설치 패키지에 서명한다.
- [ ] 개인정보 처리, 지원 연락처, 라이선스 정보를 배포 패키지에 포함한다.
