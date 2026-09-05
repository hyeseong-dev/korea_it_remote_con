"""Small native launch adapters. No shell execution or credential handling."""
import os
import platform
import plistlib
import shutil
import subprocess
import tempfile
from pathlib import Path
from .profiles import AccessError


def find_client(paths, system=None):
    system = system or platform.system()
    if system not in ("Darwin", "Windows", "Linux"):
        raise AccessError("unsupported_platform", "Supported systems: macOS, Windows and Linux.")
    override = paths.get(system)
    if override:
        candidates = [Path(override).expanduser()]
    elif system == "Darwin":
        candidates = [Path("/Applications/AnyDesk.app"), Path.home() / "Applications/AnyDesk.app"]
    else:
        found = shutil.which("AnyDesk.exe" if system == "Windows" else "anydesk")
        candidates = [Path(found)] if found else []
        if system == "Windows":
            for variable in ("ProgramFiles", "ProgramFiles(x86)", "LOCALAPPDATA"):
                if os.environ.get(variable):
                    candidates.append(Path(os.environ[variable]) / "AnyDesk/AnyDesk.exe")
    for candidate in candidates:
        if not candidate.is_absolute():
            continue
        if system == "Darwin":
            if candidate.is_dir() and (candidate / "Contents/MacOS/AnyDesk").is_file():
                return system, candidate
        elif candidate.is_file() and (system == "Windows" or os.access(candidate, os.X_OK)):
            return system, candidate
    raise AccessError("client_missing", "AnyDesk was not found; install it or configure an absolute client_paths entry.")


def address(route):
    return route.endpoint + ("/np" if route.transport == "vendor-relay" else "")


def launch(route, client, system, state_dir):
    try:
        if system == "Darwin":
            directory = Path(state_dir)
            directory.mkdir(mode=0o700, parents=True, exist_ok=True)
            if directory.is_symlink():
                raise AccessError("launch_failed", "Generated shortcut directory must not be a symbolic link.")
            os.chmod(directory, 0o700)
            fd, filename = tempfile.mkstemp(suffix=".anydeskid", prefix="connection-", dir=directory)
            with os.fdopen(fd, "wb") as output:
                plistlib.dump({"id": address(route), "type": "deskrt"}, output)
            command = ["/usr/bin/open", "-a", str(client), filename]
            subprocess.run(command, check=True, timeout=10, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        else:
            subprocess.Popen([str(client), address(route)], stdin=subprocess.DEVNULL,
                             stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    except (OSError, subprocess.SubprocessError):
        raise AccessError("launch_failed", "The AnyDesk launch request failed.") from None
