# Portable Access Implementation Plan

**Goal:** A common device-alias CLI with platform-specific AnyDesk launchers.
**Architecture:** JSON local device profiles select explicit routes. Core validates configuration and probes only Tailscale TCP routes; OS adapters launch installed clients without handling passwords.
**Tech Stack:** Python 3.9+ standard library, unittest, OS-installed AnyDesk.
**Spec:** User-approved design in conversation: device profiles, transport/provider/OS separation, application-managed authentication, no RDP changes.

## Global Constraints

- Preserve existing working scripts and local access files.
- No remote security changes, credential copying, Git commits, or uploads.
- No real endpoint, account, or password in tracked files or test fixtures.
- Use JSON rather than YAML to avoid a parser dependency.
- Local profiles: devices.local.json; ignored generated files: .local/.
- Commands: python3 remote.py list; connect ALIAS [--route NAME] [--dry-run]; doctor ALIAS [--route NAME].
- macOS: generate private plist shortcut and open with detected AnyDesk app.
- Windows/Linux: launch discovered AnyDesk executable with one validated address argument; vendor-relay uses /np.
- App paths can be overridden per OS in local settings. No shell=True.
- Report launched, not connected; only real session evidence can establish connected.
- No automatic fallback, authentication automation, or claimed runtime validation on unavailable OSes.

## Task 1: Generator — code and tests

Files: remote.py, remote_access/*.py, tests/test_remote_access.py, examples/devices.json.
Write tests for invalid profiles, option-injection endpoint rejection, missing client, primary/fallback separation, dry-run without sockets/launch/writes, Mac plist contents and Windows/Linux argument vectors. Implement small modules for profile validation, route checking, platform launching, CLI. Run `python3 -S -m unittest discover -s tests -v`.

## Task 2: Planner — local integration and documentation

Update .gitignore for local JSON, generated files and Python artifacts. Create ignored devices.local.json from existing local connection details without printing secrets. Write README with all commands, setup and platform limitations. Preserve academy.command. Save no passwords in new profiles.

## Task 3: Independent evaluator

Read this contract and inspect implementation independently. Check input validation, timeout handling, profile privacy, fallback independence and truthful statuses. Run unit tests and Git ignore checks. Fix actionable issues before final response. Actual Mac launch can be checked once; Windows/Linux only mocked until those hosts are available.
