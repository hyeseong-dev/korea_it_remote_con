package main

import (
	"context"

	"github.com/hyeseong-dev/korea_it_remote_con/desktop/internal/bridge"
)

// App is the Wails-facing boundary. Security-sensitive work stays in bridge.Manager.
type App struct {
	ctx     context.Context
	manager *bridge.Manager
}

func NewApp() *App {
	return &App{manager: bridge.NewManager()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetStatus() bridge.Status {
	return a.manager.Status(a.ctx)
}

func (a *App) SaveSettings(input bridge.SettingsInput) bridge.Status {
	return a.manager.SaveSettings(a.ctx, input)
}

func (a *App) ConnectVPN() bridge.Status {
	return a.manager.ConnectVPN(a.ctx)
}

func (a *App) DisconnectVPN() bridge.Status {
	return a.manager.DisconnectVPN(a.ctx)
}

func (a *App) ConnectAndLaunch() bridge.Status {
	return a.manager.ConnectAndLaunch(a.ctx)
}
