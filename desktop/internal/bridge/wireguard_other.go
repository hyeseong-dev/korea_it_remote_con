//go:build !windows

package bridge

import (
	"context"
	"errors"
)

var errWindowsOnly = errors.New("현재 MVP는 Windows 11에서만 터널 제어를 지원합니다.")

func tunnelState(string) (bool, bool, error)              { return false, false, errWindowsOnly }
func startTunnel(string) error                            { return errWindowsOnly }
func stopTunnel(string) error                             { return errWindowsOnly }
func installTunnel(context.Context, string, string) error { return errWindowsOnly }
func launchAnyDesk(string, string) error                  { return errWindowsOnly }
