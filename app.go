package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
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

func (a *App) WriteToHostFile() string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pkexec", "sh", "-c", "echo \"# test\" >> /etc/hosts")
	err := cmd.Run()
	if err != nil {
	}
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	return out.String() + "\n" + stderr.String()
}
