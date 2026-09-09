package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

func validInput() SettingsInput {
	return SettingsInput{
		DeviceLabel:   "Academy PC",
		TargetAddress: "academy-pc.example.ts.net",
		TailscalePath: `C:\Program Files\Tailscale\tailscale.exe`,
		AnyDeskPath:   `C:\Program Files (x86)\AnyDesk\AnyDesk.exe`,
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SettingsInput)
	}{
		{"empty target", func(i *SettingsInput) { i.TargetAddress = "" }},
		{"url target", func(i *SettingsInput) { i.TargetAddress = "https://remote.example.ts.net" }},
		{"option target", func(i *SettingsInput) { i.TargetAddress = "--help" }},
		{"loopback target", func(i *SettingsInput) { i.TargetAddress = "127.0.0.1" }},
		{"non executable", func(i *SettingsInput) { i.AnyDeskPath = `C:\AnyDesk.bat` }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput()
			test.mutate(&input)
			if err := validateConfig(configFromInput(input)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestTargetAddressAcceptsIPv4AndMagicDNS(t *testing.T) {
	for _, address := range []string{"100.64.1.2", "academy-pc", "academy-pc.example.ts.net"} {
		t.Run(address, func(t *testing.T) {
			input := validInput()
			input.TargetAddress = address
			if err := validateConfig(configFromInput(input)); err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	want := configFromInput(validInput())
	if err := saveConfig(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestLegacyWireGuardConfigMigratesInMemory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := `{
		"version": 1,
		"deviceLabel": "Academy PC",
		"vpnAddress": "100.64.1.2",
		"anyDeskPort": 7070,
		"tunnelName": "academy",
		"tunnelConfigPath": "C:\\private.conf",
		"wireGuardPath": "C:\\WireGuard\\wireguard.exe",
		"anyDeskPath": "C:\\AnyDesk\\AnyDesk.exe"
	}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 2 || got.TargetAddress != "100.64.1.2" || got.TailscalePath != "" {
		t.Fatalf("unexpected migrated config: %#v", got)
	}
}

func TestUnknownFieldsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":2,"secret":"leak"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseTailscaleState(t *testing.T) {
	tests := map[string]string{
		`{"BackendState":"Running"}`:    "running",
		`{"BackendState":"Stopped"}`:    "stopped",
		`{"BackendState":"NeedsLogin"}`: "needs-login",
		`{"BackendState":"Starting"}`:   "starting",
	}
	for input, want := range tests {
		got, err := parseTailscaleState([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", input, err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}
