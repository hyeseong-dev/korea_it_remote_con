package bridge

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Status struct {
	Configured   bool          `json:"configured"`
	VPNState     string        `json:"vpnState"`
	TargetState  string        `json:"targetState"`
	AnyDeskState string        `json:"anyDeskState"`
	Message      string        `json:"message"`
	UpdatedAt    string        `json:"updatedAt"`
	Settings     SettingsInput `json:"settings"`
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

	installed, running, err := tunnelState(config.TunnelName)
	if err != nil {
		status.VPNState = "error"
		status.Message = "WireGuard 터널 상태를 확인하지 못했습니다."
	} else if !installed {
		status.VPNState = "not-installed"
		status.Message = "WireGuard 터널이 아직 설치되지 않았습니다."
	} else if !running {
		status.VPNState = "stopped"
		status.Message = "VPN 연결이 꺼져 있습니다."
	} else {
		status.VPNState = "running"
		if probe(ctx, config.VPNAddress, config.AnyDeskPort) {
			status.TargetState = "reachable"
			status.Message = "원격 Windows PC에 연결할 준비가 됐습니다."
		} else {
			status.TargetState = "unreachable"
			status.Message = "VPN은 연결됐지만 원격 AnyDesk 포트에 응답이 없습니다."
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
	config := configFromInput(input)
	if err := saveConfig(m.configPath, config); err != nil {
		status := m.baseStatus()
		status.Message = safeError(err)
		return status
	}
	return m.Status(ctx)
}

func (m *Manager) ConnectVPN(ctx context.Context) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	installed, running, err := tunnelState(config.TunnelName)
	if err != nil {
		return m.failure(errors.New("WireGuard 터널 상태를 확인하지 못했습니다."))
	}
	if !installed {
		if config.TunnelConfigPath == "" {
			return m.failure(errors.New("처음 연결하려면 WireGuard 설정 파일 경로가 필요합니다."))
		}
		if _, err := os.Stat(config.TunnelConfigPath); err != nil {
			return m.failure(errors.New("WireGuard 설정 파일을 찾을 수 없습니다."))
		}
		wireguard, err := findWireGuard(config.WireGuardPath)
		if err != nil {
			return m.failure(err)
		}
		if err := installTunnel(ctx, wireguard, config.TunnelConfigPath); err != nil {
			return m.failure(errors.New("터널 설치에 실패했습니다. 관리자 권한으로 실행했는지 확인하세요."))
		}
	} else if !running {
		if err := startTunnel(config.TunnelName); err != nil {
			return m.failure(errors.New("VPN 시작에 실패했습니다. 관리자 권한이 필요할 수 있습니다."))
		}
	}
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		_, running, _ = tunnelState(config.TunnelName)
		if running {
			return m.Status(ctx)
		}
		time.Sleep(250 * time.Millisecond)
	}
	return m.failure(errors.New("VPN 시작 시간이 초과되었습니다."))
}

func (m *Manager) DisconnectVPN(ctx context.Context) Status {
	config, err := loadConfig(m.configPath)
	if err != nil {
		return m.failure(err)
	}
	installed, running, err := tunnelState(config.TunnelName)
	if err != nil {
		return m.failure(errors.New("WireGuard 터널 상태를 확인하지 못했습니다."))
	}
	if installed && running {
		if err := stopTunnel(config.TunnelName); err != nil {
			return m.failure(errors.New("VPN 종료에 실패했습니다. 관리자 권한이 필요할 수 있습니다."))
		}
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
	if err := launchAnyDesk(anydesk, config.VPNAddress); err != nil {
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

func findWireGuard(override string) (string, error) {
	return findExecutable(override, []string{
		filepath.Join(os.Getenv("ProgramFiles"), "WireGuard", "wireguard.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "WireGuard", "wireguard.exe"),
	}, "WireGuard")
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
