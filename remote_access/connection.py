"""Path checks; relay routes deliberately have no Tailscale dependency."""
import subprocess
import sys
from .profiles import AccessError


def probe(route):
    if route.transport == "vendor-relay":
        return "not_probed"
    # A subprocess bounds DNS resolution as well as TCP connect time.
    script = "import socket,sys; socket.create_connection((sys.argv[1],int(sys.argv[2])),timeout=3).close()"
    try:
        subprocess.run([sys.executable, "-S", "-c", script, route.endpoint, str(route.port)],
                       check=True, timeout=5, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    except (OSError, subprocess.SubprocessError):
        raise AccessError("unreachable", "TCP route probe failed or timed out; select a fallback route explicitly.") from None
    return "reachable"
