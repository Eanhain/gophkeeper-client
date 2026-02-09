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
	cfg, err := configs.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	client := clientconn.New(cfg.HTTP.Host, cfg.HTTP.Port)
	cache := storage.NewCache(cfg.Crypto.Key)
	cache.Load()

	uc := usecase.New(client, cache)

	p := tea.NewProgram(tui.New(uc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
