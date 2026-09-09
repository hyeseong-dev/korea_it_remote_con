package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const configVersion = 3

var hostnamePattern = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)

type DeviceProfile struct {
	ID            string `json:"id"`
	DeviceLabel   string `json:"deviceLabel"`
	TargetAddress string `json:"targetAddress"`
}

type Config struct {
	Version         int             `json:"version"`
	ActiveProfileID string          `json:"activeProfileId"`
	Profiles        []DeviceProfile `json:"profiles"`
	AnyDeskPort     int             `json:"anyDeskPort"`
	TailscalePath   string          `json:"tailscalePath,omitempty"`
	AnyDeskPath     string          `json:"anyDeskPath,omitempty"`
}

type SettingsInput struct {
	ProfileID     string `json:"profileId"`
	DeviceLabel   string `json:"deviceLabel"`
	TargetAddress string `json:"targetAddress"`
	TailscalePath string `json:"tailscalePath"`
	AnyDeskPath   string `json:"anyDeskPath"`
}

type legacyConfigV2 struct {
	Version       int    `json:"version"`
	DeviceLabel   string `json:"deviceLabel"`
	TargetAddress string `json:"targetAddress"`
	AnyDeskPort   int    `json:"anyDeskPort"`
	TailscalePath string `json:"tailscalePath,omitempty"`
	AnyDeskPath   string `json:"anyDeskPath,omitempty"`
}

type legacyConfigV1 struct {
	Version          int    `json:"version"`
	DeviceLabel      string `json:"deviceLabel"`
	VPNAddress       string `json:"vpnAddress"`
	AnyDeskPort      int    `json:"anyDeskPort"`
	TunnelName       string `json:"tunnelName"`
	TunnelConfigPath string `json:"tunnelConfigPath,omitempty"`
	WireGuardPath    string `json:"wireGuardPath,omitempty"`
	AnyDeskPath      string `json:"anyDeskPath,omitempty"`
}

type configEnvelope struct {
	Version int `json:"version"`
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

func newProfileID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (c Config) activeProfile() (DeviceProfile, error) {
	for _, profile := range c.Profiles {
		if profile.ID == c.ActiveProfileID {
			return profile, nil
		}
	}
	return DeviceProfile{}, configError{"선택된 원격 PC를 찾을 수 없습니다."}
}

func (c Config) settings() SettingsInput {
	profile, _ := c.activeProfile()
	return SettingsInput{ProfileID: profile.ID, DeviceLabel: profile.DeviceLabel, TargetAddress: profile.TargetAddress, TailscalePath: c.TailscalePath, AnyDeskPath: c.AnyDeskPath}
}

func validateConfig(c Config) error {
	if c.Version != configVersion {
		return configError{"지원하지 않는 설정 버전입니다."}
	}
	if len(c.Profiles) == 0 || len(c.Profiles) > 50 {
		return configError{"원격 PC는 1~50개까지 저장할 수 있습니다."}
	}
	seen := map[string]bool{}
	for _, profile := range c.Profiles {
		if len(profile.ID) != 24 || seen[profile.ID] {
			return configError{"원격 PC 식별자가 올바르지 않습니다."}
		}
		seen[profile.ID] = true
		if strings.TrimSpace(profile.DeviceLabel) == "" || len(profile.DeviceLabel) > 80 {
			return configError{"장치 이름을 1~80자로 입력하세요."}
		}
		if err := validateTargetAddress(profile.TargetAddress); err != nil {
			return err
		}
	}
	if _, err := c.activeProfile(); err != nil {
		return err
	}
	if c.AnyDeskPort != 7070 {
		return configError{"현재 버전에서는 AnyDesk 기본 포트 7070만 지원합니다."}
	}
	for label, path := range map[string]string{"Tailscale": c.TailscalePath, "AnyDesk": c.AnyDeskPath} {
		if path != "" && (!filepath.IsAbs(path) || !strings.EqualFold(filepath.Ext(path), ".exe")) {
			return configError{fmt.Sprintf("%s 실행 파일은 .exe 절대 경로여야 합니다.", label)}
		}
	}
	return nil
}

func validateTargetAddress(address string) error {
	address = strings.TrimSpace(address)
	if ip := net.ParseIP(address); ip != nil {
		if ip.To4() == nil || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() {
			return configError{"원격 주소에는 접속 가능한 IPv4 또는 MagicDNS 이름을 입력하세요."}
		}
		return nil
	}
	if len(address) == 0 || len(address) > 253 || !hostnamePattern.MatchString(address) || strings.EqualFold(address, "localhost") {
		return configError{"원격 주소에는 접속 가능한 IPv4 또는 MagicDNS 이름을 입력하세요."}
	}
	return nil
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var envelope configEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Config{}, configError{"설정 파일을 읽을 수 없습니다."}
	}
	var config Config
	switch envelope.Version {
	case configVersion:
		if err := decodeStrict(data, &config); err != nil {
			return Config{}, configError{"설정 파일을 읽을 수 없습니다."}
		}
	case 2:
		var legacy legacyConfigV2
		if err := decodeStrict(data, &legacy); err != nil {
			return Config{}, configError{"기존 설정 파일을 읽을 수 없습니다."}
		}
		config, err = migrateSingleProfile(legacy.DeviceLabel, legacy.TargetAddress, legacy.TailscalePath, legacy.AnyDeskPath)
	case 1:
		var legacy legacyConfigV1
		if err := decodeStrict(data, &legacy); err != nil {
			return Config{}, configError{"기존 설정 파일을 읽을 수 없습니다."}
		}
		config, err = migrateSingleProfile(legacy.DeviceLabel, legacy.VPNAddress, "", legacy.AnyDeskPath)
	default:
		return Config{}, configError{"지원하지 않는 설정 버전입니다."}
	}
	if err != nil {
		return Config{}, err
	}
	if err := validateConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func migrateSingleProfile(label, address, tailscalePath, anyDeskPath string) (Config, error) {
	id, err := newProfileID()
	if err != nil {
		return Config{}, err
	}
	return Config{Version: configVersion, ActiveProfileID: id, Profiles: []DeviceProfile{{ID: id, DeviceLabel: label, TargetAddress: address}}, AnyDeskPort: 7070, TailscalePath: tailscalePath, AnyDeskPath: anyDeskPath}, nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing data")
	}
	return nil
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

func isNotConfigured(err error) bool { return errors.Is(err, os.ErrNotExist) }
