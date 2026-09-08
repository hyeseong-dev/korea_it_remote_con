package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const configVersion = 1

var tunnelNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

type Config struct {
	Version          int    `json:"version"`
	DeviceLabel      string `json:"deviceLabel"`
	VPNAddress       string `json:"vpnAddress"`
	AnyDeskPort      int    `json:"anyDeskPort"`
	TunnelName       string `json:"tunnelName"`
	TunnelConfigPath string `json:"tunnelConfigPath,omitempty"`
	WireGuardPath    string `json:"wireGuardPath,omitempty"`
	AnyDeskPath      string `json:"anyDeskPath,omitempty"`
}

type SettingsInput struct {
	DeviceLabel      string `json:"deviceLabel"`
	VPNAddress       string `json:"vpnAddress"`
	TunnelName       string `json:"tunnelName"`
	TunnelConfigPath string `json:"tunnelConfigPath"`
	WireGuardPath    string `json:"wireGuardPath"`
	AnyDeskPath      string `json:"anyDeskPath"`
}

type configError struct{ message string }

func (e configError) Error() string { return e.message }

func defaultConfigPath() string {
	if override := os.Getenv("REMOTE_BRIDGE_CONFIG"); override != "" {
		return override
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "remote-bridge.local.json"
	}
	return filepath.Join(dir, "RemoteBridge", "config.json")
}

func configFromInput(input SettingsInput) Config {
	return Config{
		Version:          configVersion,
		DeviceLabel:      strings.TrimSpace(input.DeviceLabel),
		VPNAddress:       strings.TrimSpace(input.VPNAddress),
		AnyDeskPort:      7070,
		TunnelName:       strings.TrimSpace(input.TunnelName),
		TunnelConfigPath: strings.TrimSpace(input.TunnelConfigPath),
		WireGuardPath:    strings.TrimSpace(input.WireGuardPath),
		AnyDeskPath:      strings.TrimSpace(input.AnyDeskPath),
	}
}

func (c Config) settings() SettingsInput {
	return SettingsInput{
		DeviceLabel: c.DeviceLabel, VPNAddress: c.VPNAddress, TunnelName: c.TunnelName,
		TunnelConfigPath: c.TunnelConfigPath, WireGuardPath: c.WireGuardPath, AnyDeskPath: c.AnyDeskPath,
	}
}

func validateConfig(c Config) error {
	if c.Version != configVersion {
		return configError{"지원하지 않는 설정 버전입니다."}
	}
	if strings.TrimSpace(c.DeviceLabel) == "" || len(c.DeviceLabel) > 80 {
		return configError{"장치 이름을 1~80자로 입력하세요."}
	}
	ip := net.ParseIP(c.VPNAddress)
	if ip == nil || ip.To4() == nil {
		return configError{"VPN 주소에는 유효한 IPv4 주소를 입력하세요."}
	}
	if !tunnelNamePattern.MatchString(c.TunnelName) {
		return configError{"터널 이름은 영문, 숫자, 밑줄, 하이픈만 사용할 수 있습니다."}
	}
	if c.AnyDeskPort != 7070 {
		return configError{"MVP에서는 AnyDesk 기본 포트 7070만 지원합니다."}
	}
	if c.TunnelConfigPath != "" {
		if !filepath.IsAbs(c.TunnelConfigPath) {
			return configError{"WireGuard 설정 파일은 절대 경로여야 합니다."}
		}
		lower := strings.ToLower(c.TunnelConfigPath)
		if !strings.HasSuffix(lower, ".conf") && !strings.HasSuffix(lower, ".conf.dpapi") {
			return configError{"WireGuard 설정 파일은 .conf 또는 .conf.dpapi 형식이어야 합니다."}
		}
	}
	for label, path := range map[string]string{"WireGuard": c.WireGuardPath, "AnyDesk": c.AnyDeskPath} {
		if path != "" && (!filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".exe")) {
			return configError{fmt.Sprintf("%s 실행 파일은 .exe 절대 경로여야 합니다.", label)}
		}
	}
	return nil
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return Config{}, configError{"설정 파일을 읽을 수 없습니다."}
	}
	if err := validateConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func saveConfig(path string, config Config) error {
	if err := validateConfig(config); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func isNotConfigured(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
