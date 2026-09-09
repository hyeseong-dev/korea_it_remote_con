package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Status struct {
	Configured      bool            `json:"configured"`
	VPNState        string          `json:"vpnState"`
	TargetState     string          `json:"targetState"`
	AnyDeskState    string          `json:"anyDeskState"`
	Message         string          `json:"message"`
	UpdatedAt       string          `json:"updatedAt"`
	Settings        SettingsInput   `json:"settings"`
	Profiles        []DeviceProfile `json:"profiles"`
	ActiveProfileID string          `json:"activeProfileId"`
}

type Manager struct {
	configPath string
}

func NewManager() *Manager { return &Manager{configPath: defaultConfigPath()} }

func newManagerAt(path string) *Manager { return &Manager{configPath: path} }

func (m *Manager) baseStatus() Status {
	return Status{VPNState: "unknown", TargetState: "unchecked", AnyDeskState: "unknown", UpdatedAt: time.Now().Format(time.RFC3339)}
}

func (m *Manager) Status(ctx context.Context) Status {
	status := m.baseStatus()
	config, err := loadConfig(m.configPath)
	if err != nil {
		if isNotConfigured(err) {
			status.Message = "연결 정보를 먼저 설정하세요."
		} else {
			status.Message = safeError(err)
		}
		return status
	}
	status.Configured = true
	status.Settings = config.settings()
	status.Profiles = config.Profiles
	status.ActiveProfileID = config.ActiveProfileID
	profile, profileErr := config.activeProfile()
	if profileErr != nil {
		status.Message = safeError(profileErr)
		return status
	}

	tailscale, err := findTailscale(config.TailscalePath)
	if err != nil {
		status.VPNState = "missing"
		status.Message = "Tailscale을 설치하거나 실행 파일 경로를 설정하세요."
	} else {
		state, stateErr := tailscaleState(ctx, tailscale)
		if stateErr != nil {
			status.VPNState = "error"
			status.Message = "Tailscale 상태를 확인하지 못했습니다."
		} else {
			status.VPNState = state
			switch state {
			case "running":
				if probe(ctx, profile.TargetAddress, config.AnyDeskPort) {
					status.TargetState = "reachable"
					status.Message = "원격 Windows PC에 연결할 준비가 됐습니다."
				} else {
					status.TargetState = "unreachable"
					status.Message = "Tailscale은 연결됐지만 원격 AnyDesk 포트에 응답이 없습니다."
				}
			case "needs-login":
				status.Message = "Tailscale 앱에서 로그인한 후 다시 시도하세요."
			case "stopped":
				status.Message = "Tailscale 연결이 꺼져 있습니다."
			case "starting":
				status.Message = "Tailscale 연결을 준비하고 있습니다."
			}
		}
	}

	if _, err := findAnyDesk(config.AnyDeskPath); err == nil {
		status.AnyDeskState = "ready"
	} else {
		status.AnyDeskState = "missing"
		if status.Message == "" {
			status.Message = "AnyDesk를 설치하거나 실행 파일 경로를 설정하세요."
		}
	}
	return status
}

func (m *Manager) SaveSettings(ctx context.Context, input SettingsInput) Status {
	config, err := loadConfig(m.configPath)
	if err != nil && !isNotConfigured(err) {
		return m.failure(err)
	}
	if isNotConfigured(err) {
		id, idErr := newProfileID()
		if idErr != nil {
			return m.failure(idErr)
		}
		config = Config{Version: configVersion, ActiveProfileID: id, Profiles: []DeviceProfile{{ID: id}}, AnyDeskPort: 7070}
	}
	profileID := input.ProfileID
	if profileID == "" {
		id, idErr := newProfileID()
		if idErr != nil {
			return m.failure(idErr)
		}
		profileID = id
		config.Profiles = append(config.Profiles, DeviceProfile{ID: profileID})
	}
	found := false
	for index := range config.Profiles {
		if config.Profiles[index].ID == profileID {
			config.Profiles[index].DeviceLabel = strings.TrimSpace(input.DeviceLabel)
			config.Profiles[index].TargetAddress = strings.TrimSpace(input.TargetAddress)
			found = true
			break
		}
	}
	if !found {
		return m.failure(configError{"선택된 원격 PC를 찾을 수 없습니다."})
	}
	config.ActiveProfileID = profileID
	config.TailscalePath = strings.TrimSpace(input.TailscalePath)
	config.AnyDeskPath = strings.TrimSpace(input.AnyDeskPath)
	if err := saveConfig(m.configPath, config); err != nil {
		status := m.baseStatus()
		status.Message = safeError(err)
		return status
	}
	return m.Status(ctx)
}

func (m *Manager) SelectProfile(ctx context.Context, profileID string) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	config.ActiveProfileID = profileID
	if err := saveConfig(m.configPath, config); err != nil {
		return m.failure(err)
	}
	return m.Status(ctx)
}

func (m *Manager) DeleteProfile(ctx context.Context, profileID string) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	if len(config.Profiles) == 1 {
		return m.failure(configError{"마지막 원격 PC는 삭제할 수 없습니다."})
	}
	profiles := make([]DeviceProfile, 0, len(config.Profiles)-1)
	found := false
	for _, profile := range config.Profiles {
		if profile.ID == profileID {
			found = true
			continue
		}
		profiles = append(profiles, profile)
	}
	if !found {
		return m.failure(configError{"선택된 원격 PC를 찾을 수 없습니다."})
	}
	config.Profiles = profiles
	if config.ActiveProfileID == profileID {
		config.ActiveProfileID = profiles[0].ID
	}
	if err := saveConfig(m.configPath, config); err != nil {
		return m.failure(err)
	}
	return m.Status(ctx)
}

func (m *Manager) ConnectVPN(ctx context.Context) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	tailscale, err := findTailscale(config.TailscalePath)
	if err != nil {
		return m.failure(err)
	}
	state, err := tailscaleState(ctx, tailscale)
	if err != nil {
		return m.failure(errors.New("Tailscale 상태를 확인하지 못했습니다."))
	}
	if state == "needs-login" {
		status := m.Status(ctx)
		status.Message = "Tailscale 앱에서 로그인한 후 다시 시도하세요."
		return status
	}
	if state != "running" {
		connectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := connectTailscale(connectCtx, tailscale); err != nil {
			return m.failure(errors.New("Tailscale 연결에 실패했습니다. Tailscale 앱의 로그인 상태를 확인하세요."))
		}
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		state, _ = tailscaleState(ctx, tailscale)
		if state == "running" {
			return m.Status(ctx)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return m.failure(errors.New("Tailscale 연결 시간이 초과되었습니다."))
}

func (m *Manager) DisconnectVPN(ctx context.Context) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	tailscale, err := findTailscale(config.TailscalePath)
	if err != nil {
		return m.failure(err)
	}
	if err := disconnectTailscale(ctx, tailscale); err != nil {
		return m.failure(errors.New("Tailscale 연결 종료에 실패했습니다."))
	}
	return m.Status(ctx)
}

func (m *Manager) ConnectAndLaunch(ctx context.Context) Status {
	status := m.ConnectVPN(ctx)
	if status.VPNState != "running" || status.TargetState != "reachable" {
		return status
	}
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	anydesk, err := findAnyDesk(config.AnyDeskPath)
	if err != nil {
		return m.failure(err)
	}
	profile, err := config.activeProfile()
	if err != nil {
		return m.failure(err)
	}
	if err := launchAnyDesk(anydesk, profile.TargetAddress); err != nil {
		return m.failure(errors.New("AnyDesk 실행 요청에 실패했습니다."))
	}
	status = m.Status(ctx)
	status.Message = "AnyDesk를 실행했습니다. 앱에서 인증을 완료하세요."
	return status
}

func (m *Manager) failure(err error) Status {
	status := m.baseStatus()
	status.Message = safeError(err)
	if config, loadErr := loadConfig(m.configPath); loadErr == nil {
		status.Configured = true
		status.Settings = config.settings()
		status.Profiles = config.Profiles
		status.ActiveProfileID = config.ActiveProfileID
	}
	return status
}

func safeError(err error) string {
	var validation configError
	if errors.As(err, &validation) {
		return validation.Error()
	}
	return err.Error()
}

func probe(ctx context.Context, host string, port int) bool {
	dialer := net.Dialer{Timeout: 3 * time.Second}
	connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, fmt.Sprint(port)))
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func findTailscale(override string) (string, error) {
	return findExecutable(override, []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Tailscale", "tailscale.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Tailscale", "tailscale.exe"),
	}, "Tailscale")
}

func findAnyDesk(override string) (string, error) {
	return findExecutable(override, []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "AnyDesk", "AnyDesk.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "AnyDesk", "AnyDesk.exe"),
	}, "AnyDesk")
}

func findExecutable(override string, candidates []string, label string) (string, error) {
	if override != "" {
		candidates = []string{override}
	}
	for _, candidate := range candidates {
		if candidate == "" || strings.HasPrefix(candidate, string(filepath.Separator)) {
			continue
		}
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s를 설치하거나 실행 파일 경로를 설정하세요.", label)
}

type tailscaleStatus struct {
	BackendState string `json:"BackendState"`
}

func parseTailscaleState(data []byte) (string, error) {
	var status tailscaleStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return "", err
	}
	switch strings.ToLower(status.BackendState) {
	case "running":
		return "running", nil
	case "stopped":
		return "stopped", nil
	case "needslogin", "nostate":
		return "needs-login", nil
	case "starting":
		return "starting", nil
	default:
		return "", fmt.Errorf("unknown Tailscale state: %q", status.BackendState)
	}
}
