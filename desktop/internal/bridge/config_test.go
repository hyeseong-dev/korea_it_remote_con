package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

func validInput() SettingsInput {
	return SettingsInput{
		DeviceLabel: "Academy PC", VPNAddress: "10.88.0.2", TunnelName: "academy",
		TunnelConfigPath: `C:\ProgramData\RemoteBridge\academy.conf.dpapi`,
		WireGuardPath:    `C:\Program Files\WireGuard\wireguard.exe`,
		AnyDeskPath:      `C:\Program Files (x86)\AnyDesk\AnyDesk.exe`,
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SettingsInput)
	}{
		{"bad IP", func(i *SettingsInput) { i.VPNAddress = "host;command" }},
		{"bad tunnel", func(i *SettingsInput) { i.TunnelName = "--delete" }},
		{"relative config", func(i *SettingsInput) { i.TunnelConfigPath = "private.conf" }},
		{"wrong config extension", func(i *SettingsInput) { i.TunnelConfigPath = `C:\private.txt` }},
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
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" {
		t.Fatal("empty config")
	}
}

func TestUnknownFieldsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"secret":"leak"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("expected parse error")
	}
}
