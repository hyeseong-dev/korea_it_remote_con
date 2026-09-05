#!/bin/zsh
# Connection details stay in the ignored local AnyDesk shortcut.
set -eu
base_dir=${0:A:h}
shortcut="$base_dir/academy.anydeskid"

if [[ ! -d /Applications/AnyDesk.app ]]; then
  print -u2 'AnyDesk가 설치되어 있지 않습니다.'
  exit 1
fi
if [[ ! -r "$shortcut" ]]; then
  print -u2 'academy.anydeskid 파일이 없습니다. README.md의 로컬 설정 절차를 확인하세요.'
  exit 1
fi
if ! address=$(/usr/bin/plutil -extract id raw -o - "$shortcut" 2>/dev/null); then
  print -u2 '바로가기의 접속 주소를 읽을 수 없습니다. academy.anydeskid를 확인하세요.'
  exit 1
fi
if [[ -z "$address" || "$address" == YOUR_TAILSCALE_IP || "$address" == -* ]]; then
  print -u2 '바로가기에 실제 Tailscale IP를 설정하세요.'
  exit 1
fi
if ! /usr/bin/nc -G 5 -z "$address" 7070 2>/dev/null; then
  print -u2 '학원 PC에 연결되지 않습니다. Mac의 Tailscale 연결과 학원 PC 전원을 확인하세요.'
  exit 1
fi
print '학원 PC 연결을 엽니다. 암호 요청 시 로컬에 보관한 AnyDesk 암호를 사용하세요.'
/usr/bin/open -a AnyDesk "$shortcut"
