"""Strict local profile validation; never include profile values in errors."""
import json
import re
from dataclasses import dataclass
from pathlib import Path


class AccessError(Exception):
    def __init__(self, status, message):
        super().__init__(message)
        self.status = status


@dataclass(frozen=True)
class Route:
    name: str
    provider: str
    transport: str
    endpoint: str
    port: int = 7070


@dataclass(frozen=True)
class Device:
    label: str
    routes: tuple


@dataclass(frozen=True)
class Profile:
    devices: dict
    client_paths: dict


def invalid(message):
    raise AccessError("invalid_config", message)


def mapping(value, allowed, required):
    if not isinstance(value, dict) or set(value) - set(allowed) or not set(required) <= set(value):
        invalid("Configuration has missing or unsupported fields.")


def name(value):
    return isinstance(value, str) and re.fullmatch(r"[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}", value) is not None


def endpoint_valid(value):
    # Allow plain DNS/IPv4/AnyDesk addresses only. Options, ports, paths and
    # credential-bearing URLs must never be passed through to the client.
    return (isinstance(value, str) and len(value) <= 253
            and all(re.fullmatch(r"[a-zA-Z0-9](?:[a-zA-Z0-9_-]{0,61}[a-zA-Z0-9])?", part)
                    for part in value.split(".")))


def parse_profile(data):
    mapping(data, ("version", "devices", "client_paths"), ("version", "devices"))
    if type(data["version"]) is not int or data["version"] != 1:
        invalid("Only profile version 1 is supported.")
    if not isinstance(data["devices"], dict) or not data["devices"]:
        invalid("At least one device is required.")
    paths = data.get("client_paths", {})
    mapping(paths, ("Darwin", "Windows", "Linux"), ())
    if any(not isinstance(path, str) or not path or "\x00" in path for path in paths.values()):
        invalid("Client paths must be nonempty strings.")
    devices = {}
    for alias, entry in data["devices"].items():
        if not name(alias):
            invalid("Device aliases must use letters, digits, underscores or hyphens.")
        mapping(entry, ("label", "routes", "authentication"), ("label", "routes", "authentication"))
        if not isinstance(entry["label"], str) or not entry["label"].strip():
            invalid("Device labels must be nonempty strings.")
        if entry["authentication"] != {"mode": "application-managed"}:
            invalid("Only application-managed authentication is supported; do not store credentials here.")
        if not isinstance(entry["routes"], list) or not entry["routes"]:
            invalid("Every device must have at least one route.")
        routes = []
        seen = set()
        for route in entry["routes"]:
            mapping(route, ("name", "provider", "transport", "endpoint", "port"),
                    ("name", "provider", "transport", "endpoint"))
            if not name(route["name"]) or route["name"] in seen:
                invalid("Route names must be valid and unique within each device.")
            seen.add(route["name"])
            if route["provider"] != "anydesk" or route["transport"] not in ("tailscale", "vendor-relay"):
                invalid("Unsupported provider or transport.")
            if not endpoint_valid(route["endpoint"]):
                invalid("Endpoint must be a plain hostname, IPv4 address or AnyDesk ID.")
            port = route.get("port", 7070)
            if type(port) is not int or not 1 <= port <= 65535:
                invalid("Port must be an integer from 1 to 65535.")
            if port != 7070:
                invalid("This adapter supports only the default AnyDesk TCP port 7070.")
            if route["transport"] == "vendor-relay" and "port" in route:
                invalid("Vendor relay routes do not accept a TCP probe port.")
            routes.append(Route(**route))
        devices[alias] = Device(entry["label"], tuple(routes))
    return Profile(devices, paths)


def load_profile(path):
    try:
        return parse_profile(json.loads(Path(path).read_text(encoding="utf-8")))
    except (OSError, UnicodeError, json.JSONDecodeError):
        raise AccessError("invalid_config", "Cannot read a valid UTF-8 JSON profile.") from None


def select_route(profile, alias, route_name=None):
    device = profile.devices.get(alias)
    if device is None:
        raise AccessError("unknown_device", "Device alias is not in this profile.")
    if route_name is None:
        return device.routes[0]
    for route in device.routes:
        if route.name == route_name:
            return route
    raise AccessError("unknown_route", "Route is not configured for this device.")
