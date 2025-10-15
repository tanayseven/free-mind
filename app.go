package main

import (
	"context"
	"fmt"
	"runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Read(name string) string {
	return "Read " + name + " file successfully"
}

func (a *App) Write(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) HostsFilePath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts`
	case "darwin", "linux":
		return "/etc/hosts"
	default:
		return ""
	}
}
