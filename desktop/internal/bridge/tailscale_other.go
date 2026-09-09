//go:build !windows

package bridge

import (
	"context"
	"errors"
)

var errWindowsOnly = errors.New("현재 버전은 Windows 11에서만 연결 제어를 지원합니다.")

func tailscaleState(context.Context, string) (string, error) { return "", errWindowsOnly }
func connectTailscale(context.Context, string) error         { return errWindowsOnly }
func disconnectTailscale(context.Context, string) error      { return errWindowsOnly }
func launchAnyDesk(string, string) error                     { return errWindowsOnly }
