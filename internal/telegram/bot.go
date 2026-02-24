package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"perezvonish/factorio-server-manager/internal/domain"
	"perezvonish/factorio-server-manager/internal/factorio/saves"
	"perezvonish/factorio-server-manager/internal/factorio/status"
	"perezvonish/factorio-server-manager/internal/password"
)

// Bot is the Telegram bot that manages the Factorio server
type Bot struct {
	api          *tgbotapi.BotAPI
	allowedUsers map[int64]struct{}
	rcon         domain.RconExecutor
	container    domain.ContainerManager
	saves        *saves.Manager
	status       *status.Checker
	passwords    *password.Manager
}

// Config holds all dependencies needed to build a Bot
type Config struct {
	Token        string
	AllowedUsers map[int64]struct{}
	Rcon         domain.RconExecutor
	Container    domain.ContainerManager
	Saves        *saves.Manager
	Status       *status.Checker
	PasswordMgr  *password.Manager
}

func NewBot(cfg Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:          api,
		allowedUsers: cfg.AllowedUsers,
		rcon:         cfg.Rcon,
		container:    cfg.Container,
		saves:        cfg.Saves,
		status:       cfg.Status,
		passwords:    cfg.PasswordMgr,
	}, nil
}

// Start begins long-polling for Telegram updates
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	log.Printf("Бот запущен: @%s", b.api.Self.UserName)

	for update := range updates {
		b.handleUpdate(update)
	}
}

func (b *Bot) isAllowedUser(userID int64) bool {
	_, ok := b.allowedUsers[userID]
	return ok
}

func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("reply error: %v", err)
	}
}

func (b *Bot) replyDocument(chatID int64, name string, data []byte) {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name: name, Bytes: data})
	if _, err := b.api.Send(doc); err != nil {
		log.Printf("replyDocument error: %v", err)
	}
}
