//go:build windows

package bridge

import (
	"context"
	"os/exec"
)

func tailscaleState(ctx context.Context, executable string) (string, error) {
	command := exec.CommandContext(ctx, executable, "status", "--json")
	command.SysProcAttr = hiddenProcessAttributes()
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return parseTailscaleState(output)
}

func connectTailscale(ctx context.Context, executable string) error {
	command := exec.CommandContext(ctx, executable, "up", "--timeout=10s")
	command.SysProcAttr = hiddenProcessAttributes()
	return command.Run()
}

func disconnectTailscale(ctx context.Context, executable string) error {
	command := exec.CommandContext(ctx, executable, "down")
	command.SysProcAttr = hiddenProcessAttributes()
	return command.Run()
}

func launchAnyDesk(executable, address string) error {
	command := exec.Command(executable, address)
	command.SysProcAttr = hiddenProcessAttributes()
	return command.Start()
}
