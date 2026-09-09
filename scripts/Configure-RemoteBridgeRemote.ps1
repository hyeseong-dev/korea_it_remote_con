[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

if (-not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw '관리자 PowerShell에서 이 스크립트를 실행하세요.'
}

$tailscalePath = Join-Path $env:ProgramFiles 'Tailscale\tailscale.exe'
$anyDeskCandidates = @(
    (Join-Path ${env:ProgramFiles(x86)} 'AnyDesk\AnyDesk.exe'),
    (Join-Path $env:ProgramFiles 'AnyDesk\AnyDesk.exe'),
    (Join-Path $env:LOCALAPPDATA 'AnyDesk\AnyDesk.exe')
)
$anyDeskPath = $anyDeskCandidates | Where-Object { $_ -and (Test-Path -LiteralPath $_) } | Select-Object -First 1

if (-not (Test-Path -LiteralPath $tailscalePath)) { throw 'Tailscale for Windows를 먼저 설치하세요.' }
if (-not $anyDeskPath) { throw 'AnyDesk를 먼저 설치하세요.' }

$ruleName = 'RemoteBridge-AnyDesk-7070'
if (-not (Get-NetFirewallRule -Name $ruleName -ErrorAction SilentlyContinue)) {
    New-NetFirewallRule -Name $ruleName -DisplayName 'RemoteBridge AnyDesk over Tailscale' -Description 'Allow AnyDesk direct TCP only from Tailscale IPv4 addresses.' -Direction Inbound -Action Allow -Protocol TCP -LocalPort 7070 -RemoteAddress '100.64.0.0/10' -Profile Any | Out-Null
}

& $tailscalePath status --json | Out-Null
Write-Host '원격 PC 준비가 완료되었습니다.' -ForegroundColor Green
Write-Host '다음으로 AnyDesk에서 무인 접속과 직접 연결을 설정하세요.' -ForegroundColor Yellow
