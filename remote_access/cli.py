import argparse
from pathlib import Path
from .profiles import AccessError, load_profile, select_route
from .platforms import find_client, launch
from .connection import probe

DEFAULT_CONFIG = Path(__file__).resolve().parent.parent / "devices.local.json"


def main(argv=None):
    parser = argparse.ArgumentParser(description="Portable device-alias remote access; authentication stays in AnyDesk.")
    parser.add_argument("--config", type=Path, default=DEFAULT_CONFIG)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("list", help="List aliases and route names (no addresses)")
    for command in ("connect", "doctor"):
        child = commands.add_parser(command)
        child.add_argument("alias")
        child.add_argument("--route")
        if command == "connect":
            child.add_argument("--dry-run", action="store_true", help="Validate and discover client; no probing, writes or launch")
    args = parser.parse_args(argv)
    try:
        profile = load_profile(args.config)
        if args.command == "list":
            for alias, device in profile.devices.items():
                print(f"{alias}: " + ", ".join(route.name for route in device.routes))
            return 0
        route = select_route(profile, args.alias, args.route)
        system, client = find_client(profile.client_paths)
        if args.command == "connect" and args.dry_run:
            print(f"dry_run: {args.alias} / {route.name} / {route.provider} / {route.transport} / {system}; client found; connection not tested")
            return 0
        status = probe(route)
        if args.command == "doctor":
            print(f"{status}: client found; " + ("TCP path reachable; authentication and session unverified" if status == "reachable" else "relay availability and authentication must be checked in AnyDesk"))
            return 0
        launch(route, client, system, args.config.resolve().parent / ".local")
        print("launched: AnyDesk launch requested; confirm authentication and connection in the application")
        return 0
    except AccessError as error:
        print(f"{error.status}: {error}")
        return 1
