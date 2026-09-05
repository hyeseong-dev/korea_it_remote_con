import contextlib
import copy
import io
import json
import plistlib
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from remote_access import cli
from remote_access.connection import probe
from remote_access.platforms import find_client, launch
from remote_access.profiles import AccessError, parse_profile, select_route


def fixture():
    return {"version": 1, "devices": {"academy": {
        "label": "Test PC", "authentication": {"mode": "application-managed"},
        "routes": [
            {"name": "primary", "provider": "anydesk", "transport": "tailscale", "endpoint": "pc.example.invalid"},
            {"name": "fallback", "provider": "anydesk", "transport": "vendor-relay", "endpoint": "123456789"},
        ]}}}


class ProfileTests(unittest.TestCase):
    def test_select_and_missing_routes(self):
        profile = parse_profile(fixture())
        self.assertEqual(select_route(profile, "academy").name, "primary")
        self.assertEqual(select_route(profile, "academy", "fallback").transport, "vendor-relay")
        for alias, route in (("absent", None), ("academy", "absent")):
            with self.assertRaises(AccessError):
                select_route(profile, alias, route)

    def test_invalid_endpoints(self):
        for endpoint in ("--password", "-x", "123/np", "host;command", "host\narg", "a b", "https://host", "", "::1", "host:7070"):
            with self.subTest(endpoint=endpoint):
                data = fixture()
                data["devices"]["academy"]["routes"][0]["endpoint"] = endpoint
                with self.assertRaises(AccessError) as error:
                    parse_profile(data)
                self.assertEqual(error.exception.status, "invalid_config")

    def test_invalid_structure_and_credentials(self):
        bad = [None, [], {}, {"version": True, "devices": {}}]
        for field, value in (("authentication", {"mode": "application-managed", "password": "fake"}),
                             ("routes", []), ("label", None)):
            data = fixture()
            data["devices"]["academy"][field] = value
            bad.append(data)
        duplicate = fixture()
        duplicate["devices"]["academy"]["routes"][1]["name"] = "primary"
        bad.append(duplicate)
        for data in bad:
            with self.subTest(data=data), self.assertRaises(AccessError):
                parse_profile(data)

    def test_custom_ports_rejected(self):
        for port in (True, 0, 65536, "7070", 7071):
            data = fixture()
            data["devices"]["academy"]["routes"][0]["port"] = port
            with self.assertRaises(AccessError):
                parse_profile(data)


class ConnectionTests(unittest.TestCase):
    def setUp(self):
        self.profile = parse_profile(fixture())

    @patch("remote_access.connection.subprocess.run")
    def test_fallback_never_probes(self, run):
        self.assertEqual(probe(select_route(self.profile, "academy", "fallback")), "not_probed")
        run.assert_not_called()

    @patch("remote_access.connection.subprocess.run")
    def test_probe_timeout_is_bounded_and_private(self, run):
        run.side_effect = subprocess.TimeoutExpired("secret endpoint", 5)
        with self.assertRaises(AccessError) as error:
            probe(select_route(self.profile, "academy"))
        self.assertEqual(run.call_args.kwargs["timeout"], 5)
        self.assertEqual(error.exception.status, "unreachable")
        self.assertNotIn("secret", str(error.exception))

    @patch("remote_access.connection.subprocess.run")
    def test_success_probe(self, run):
        self.assertEqual(probe(select_route(self.profile, "academy")), "reachable")
        self.assertEqual(run.call_args.args[0][-1], "7070")


class PlatformTests(unittest.TestCase):
    def test_missing_override(self):
        with self.assertRaises(AccessError) as error:
            find_client({"Linux": "/nonexistent-test-client"}, "Linux")
        self.assertEqual(error.exception.status, "client_missing")

    @patch("remote_access.platforms.subprocess.Popen")
    def test_windows_linux_argv(self, popen):
        profile = parse_profile(fixture())
        for system in ("Windows", "Linux"):
            launch(select_route(profile, "academy"), "/client", system, "/unused")
            self.assertEqual(popen.call_args.args[0], ["/client", "pc.example.invalid"])
            launch(select_route(profile, "academy", "fallback"), "/client", system, "/unused")
            self.assertEqual(popen.call_args.args[0], ["/client", "123456789/np"])
            self.assertNotIn("shell", popen.call_args.kwargs)

    @patch("remote_access.platforms.subprocess.run")
    def test_macos_private_plist(self, run):
        with tempfile.TemporaryDirectory() as directory:
            launch(select_route(parse_profile(fixture()), "academy", "fallback"),
                   "/Applications/AnyDesk.app", "Darwin", Path(directory) / ".local")
            command = run.call_args.args[0]
            self.assertEqual(command[:3], ["/usr/bin/open", "-a", "/Applications/AnyDesk.app"])
            path = Path(command[3])
            self.assertEqual(plistlib.loads(path.read_bytes()), {"id": "123456789/np", "type": "deskrt"})
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)


class CLITests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.path = Path(self.temporary.name) / "profile.json"
        self.path.write_text(json.dumps(fixture()))

    @patch("remote_access.cli.launch")
    @patch("remote_access.cli.probe")
    @patch("remote_access.cli.find_client", return_value=("Darwin", Path("/Applications/AnyDesk.app")))
    def test_dry_run_no_write_probe_or_launch(self, client, probe_mock, launch_mock):
        before = list(Path(self.temporary.name).iterdir())
        with patch("socket.create_connection") as socket_mock, patch("pathlib.Path.write_text") as write, contextlib.redirect_stdout(io.StringIO()) as output:
            self.assertEqual(cli.main(["--config", str(self.path), "connect", "academy", "--dry-run"]), 0)
            write.assert_not_called()
            socket_mock.assert_not_called()
        probe_mock.assert_not_called()
        launch_mock.assert_not_called()
        self.assertEqual(list(Path(self.temporary.name).iterdir()), before)
        self.assertNotIn("pc.example.invalid", output.getvalue())

    @patch("remote_access.cli.launch")
    @patch("remote_access.cli.probe", side_effect=AccessError("unreachable", "Route unavailable"))
    @patch("remote_access.cli.find_client", return_value=("Linux", Path("/client")))
    def test_failed_primary_does_not_launch_or_fallback(self, client, check, launch_mock):
        with contextlib.redirect_stdout(io.StringIO()):
            self.assertEqual(cli.main(["--config", str(self.path), "connect", "academy"]), 1)
        self.assertEqual(check.call_count, 1)
        launch_mock.assert_not_called()

    @patch("remote_access.cli.find_client", return_value=("Linux", Path("/client")))
    def test_relay_doctor_does_not_claim_connected(self, client):
        with contextlib.redirect_stdout(io.StringIO()) as output:
            self.assertEqual(cli.main(["--config", str(self.path), "doctor", "academy", "--route", "fallback"]), 0)
        self.assertIn("not_probed", output.getvalue())
        self.assertNotIn("connected", output.getvalue())

    def test_invalid_json_is_sanitized(self):
        self.path.write_text('{"secret": BROKEN')
        with contextlib.redirect_stdout(io.StringIO()) as output:
            self.assertEqual(cli.main(["--config", str(self.path), "list"]), 1)
        self.assertNotIn("secret", output.getvalue())


if __name__ == "__main__":
    unittest.main()
