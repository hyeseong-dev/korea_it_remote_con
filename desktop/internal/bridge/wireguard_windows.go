//go:build windows

package bridge

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"golang.org/x/sys/windows"
)

func serviceName(tunnelName string) string { return "WireGuardTunnel$" + tunnelName }

func tunnelState(tunnelName string) (installed bool, running bool, err error) {
	service, err := openService(tunnelName, windows.SERVICE_QUERY_STATUS)
	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	defer windows.CloseServiceHandle(service)
	status, err := queryService(service)
	if err != nil {
		return true, false, err
	}
	return true, status == windows.SERVICE_RUNNING, nil
}

func startTunnel(tunnelName string) error {
	service, err := openService(tunnelName, windows.SERVICE_START|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(service)
	return windows.StartService(service, 0, nil)
}

func stopTunnel(tunnelName string) error {
	service, err := openService(tunnelName, windows.SERVICE_STOP|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return err
	}
	defer windows.CloseServiceHandle(service)
	var status windows.SERVICE_STATUS
	err = windows.ControlService(service, windows.SERVICE_CONTROL_STOP, &status)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		state, queryErr := queryService(service)
		if queryErr != nil {
			return queryErr
		}
		if state == windows.SERVICE_STOPPED {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return errors.New("service stop timeout")
}

func openService(tunnelName string, access uint32) (windows.Handle, error) {
	manager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return 0, err
	}
	defer windows.CloseServiceHandle(manager)
	name, err := windows.UTF16PtrFromString(serviceName(tunnelName))
	if err != nil {
		return 0, err
	}
	return windows.OpenService(manager, name, access)
}

func queryService(service windows.Handle) (uint32, error) {
	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(service, &status); err != nil {
		return 0, err
	}
	return status.CurrentState, nil
}

func installTunnel(ctx context.Context, executable, configPath string) error {
	command := exec.CommandContext(ctx, executable, "/installtunnelservice", configPath)
	command.SysProcAttr = hiddenProcessAttributes()
	return command.Run()
}

func launchAnyDesk(executable, address string) error {
	command := exec.Command(executable, address)
	command.SysProcAttr = hiddenProcessAttributes()
	return command.Start()
}
