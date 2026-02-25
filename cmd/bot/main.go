package main

import (
	"fmt"
	"log"
	"strings"

	"perezvonish/factorio-server-manager/internal/config"
	"perezvonish/factorio-server-manager/internal/docker"
	"perezvonish/factorio-server-manager/internal/factorio/mods"
	rconClient "perezvonish/factorio-server-manager/internal/factorio/rcon"
	"perezvonish/factorio-server-manager/internal/factorio/saves"
	"perezvonish/factorio-server-manager/internal/factorio/settings"
	"perezvonish/factorio-server-manager/internal/factorio/status"
	"perezvonish/factorio-server-manager/internal/password"
	"perezvonish/factorio-server-manager/internal/telegram"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Generate a fresh RCON password on every startup and persist it to disk.
	// The Factorio container reads the rconpw file when it starts
	pwManager := password.NewManager(cfg.FactorioServer.RconPwFile)
	if err := pwManager.Generate(24); err != nil {
		log.Fatalf("password: %v", err)
	}

	if err := settings.UpdatePasswords(cfg.FactorioServer.ServerSettingsFile, pwManager.Get()); err != nil {
		log.Fatalf("server settings: %v", err)
	}

	rcon := rconClient.NewClient(
		cfg.FactorioServer.RconHost,
		cfg.FactorioServer.RconPort,
		pwManager,
	)

	dockerMgr := docker.NewManager(cfg.Docker.ContainerName)
	saveMgr := saves.NewManager(cfg.FactorioServer.SavesDir)
	statusChecker := status.NewChecker(cfg.FactorioServer.GameHost, cfg.FactorioServer.GamePort)
	modsMgr := mods.NewManager(
		cfg.ModPortal.ModsDir,
		cfg.ModPortal.ModListFile,
		cfg.ModPortal.Username,
		cfg.ModPortal.Token,
		cfg.ModPortal.FactorioVersion,
	)

	bot, err := telegram.NewBot(telegram.Config{
		Token:        cfg.Telegram.BotToken,
		AllowedUsers: parseAllowedUsers(cfg.Telegram.AllowedUsers),
		Rcon:         rcon,
		Container:    dockerMgr,
		Saves:        saveMgr,
		Status:       statusChecker,
		PasswordMgr:  pwManager,
		Mods:         modsMgr,
	})
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}

	bot.Start()
}

func parseAllowedUsers(s string) map[int64]struct{} {
	users := make(map[int64]struct{})
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var id int64
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil && id != 0 {
			users[id] = struct{}{}
		}
	}
	return users
}
