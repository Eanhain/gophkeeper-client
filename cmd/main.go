// GophKeeper CLI client.
//
// The client provides a terminal user interface (TUI) for managing
// secrets stored on a GophKeeper server. It supports:
//   - registration and login via JWT
//   - CRUD operations on four secret types
//   - encrypted local SQLite cache for offline access
//   - retry logic with exponential backoff for network resilience
//
// Configuration is loaded from .env files, then overridden by CLI flags.
// Run with -h to see available flags, or -v to print the version.
package main

import (
	"fmt"
	"os"

	"github.com/Eanhain/gophkeeper-client/configs"
	clientconn "github.com/Eanhain/gophkeeper-client/internal/clientConn"
	"github.com/Eanhain/gophkeeper-client/internal/storage"
	"github.com/Eanhain/gophkeeper-client/internal/tui"
	"github.com/Eanhain/gophkeeper-client/internal/usecase"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load config: .env → env vars → CLI flags.
	cfg, err := configs.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if cfg == nil {
		// cfg is nil when -v flag was used (version already printed).
		return
	}

	// Wire up the dependency graph: HTTP client → cache → usecase → TUI.
	client := clientconn.New(cfg.HTTP.Host, cfg.HTTP.Port)
	cache := storage.NewCache(cfg.Crypto.Key)
	if err := cache.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	uc := usecase.New(client, cache)

	p := tea.NewProgram(tui.New(uc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
