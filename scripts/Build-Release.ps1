[CmdletBinding()]
param(
    [string]$Version = '3.0.0-beta.1'
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$desktopRoot = Join-Path $projectRoot 'desktop'
$releaseRoot = Join-Path $projectRoot 'release'
$packageRoot = Join-Path $releaseRoot "RemoteBridge-$Version"
$zipPath = Join-Path $releaseRoot "RemoteBridge-$Version-windows-amd64.zip"
$goBin = Join-Path $projectRoot '.local\go-toolchain\go\bin'
$wails = Join-Path $projectRoot '.local\go-bin\wails.exe'

if (-not (Test-Path -LiteralPath (Join-Path $goBin 'go.exe'))) { throw 'Go 도구를 찾을 수 없습니다.' }
if (-not (Test-Path -LiteralPath $wails)) { throw 'Wails 도구를 찾을 수 없습니다.' }

$env:Path = "$goBin;$env:Path"
Push-Location $desktopRoot
try {
    & (Join-Path $goBin 'go.exe') test ./...
    npm --prefix frontend run build
    & $wails build -clean -platform windows/amd64 -webview2 embed
} finally { Pop-Location }

if (Test-Path -LiteralPath $packageRoot) { Remove-Item -LiteralPath $packageRoot -Recurse -Force }
New-Item -ItemType Directory -Path $packageRoot -Force | Out-Null
Copy-Item -LiteralPath (Join-Path $desktopRoot 'build\bin\RemoteBridge.exe') -Destination $packageRoot
Copy-Item -LiteralPath (Join-Path $projectRoot 'README.md') -Destination $packageRoot
Copy-Item -LiteralPath (Join-Path $projectRoot 'docs\USER_SETUP.md') -Destination $packageRoot
if (Test-Path -LiteralPath $zipPath) { Remove-Item -LiteralPath $zipPath -Force }
Compress-Archive -Path (Join-Path $packageRoot '*') -DestinationPath $zipPath
Get-FileHash -LiteralPath $zipPath -Algorithm SHA256 | Format-List
