package telegram

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	if !b.isAllowedUser(userID) {
		b.reply(chatID, "⛔ Нет доступа")
		return
	}

	if update.Message.Document != nil {
		b.handleUploadSave(chatID, update.Message.Document)
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	args := update.Message.CommandArguments()

	switch update.Message.Command() {

	case "start", "help":
		b.handleHelp(chatID)

	case "status":
		b.handleStatus(chatID)

	case "players":
		b.handlePlayers(chatID)

	case "cmd":
		b.handleCmd(chatID, args)

	case "msg":
		b.handleMsg(chatID, args)

	case "save":
		b.handleSave(chatID)

	case "time":
		b.handleTime(chatID)

	case "evolution":
		b.handleEvolution(chatID)

	case "restart":
		b.handleRestart(chatID)

	case "stop":
		b.handleStopServer(chatID)

	case "startServer":
		b.handleStartServer(chatID)

	case "getPassword":
		b.handleGetPassword(chatID)

	case "uploadSave":
		b.reply(chatID, "📤 Отправь zip-файл сохранения в этот чат.")

	case "downloadSave":
		b.handleDownloadSave(chatID)
	}
}

// ── help ─────────────────────────────────────────────────────────────────────

func (b *Bot) handleHelp(chatID int64) {
	b.reply(chatID, `🏭 Factorio Bot

/status — статус сервера
/players — игроки онлайн
/cmd <команда> — RCON команда
/msg <текст> — сообщение в чат игры
/save — принудительное сохранение
/time — время в игре
/evolution — уровень эволюции
/restart — перезапустить через RCON (Docker restart policy)

/stop — полностью остановить контейнер
/startServer — запустить контейнер

/getPassword — пароль RCON подключения
/downloadSave — скачать текущее сохранение
/uploadSave — загрузить сохранение (или просто отправь zip-файл)`)
}

// ── server status ─────────────────────────────────────────────────────────────

func (b *Bot) handleStatus(chatID int64) {
	b.reply(chatID, fmt.Sprintf("Сервер: %s", b.status.Check()))
}

// ── players ───────────────────────────────────────────────────────────────────

func (b *Bot) handlePlayers(chatID int64) {
	resp, err := b.rcon.Execute("/players online")
	if err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	if strings.TrimSpace(resp) == "" {
		b.reply(chatID, "😴 Нет игроков онлайн")
	} else {
		b.reply(chatID, "👥 "+resp)
	}
}

// ── cmd ───────────────────────────────────────────────────────────────────────

func (b *Bot) handleCmd(chatID int64, args string) {
	if args == "" {
		b.reply(chatID, "Использование: /cmd <команда>")
		return
	}
	resp, err := b.rcon.Execute(args)
	if err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	if strings.TrimSpace(resp) == "" {
		b.reply(chatID, "✅ Выполнено")
	} else {
		b.reply(chatID, resp)
	}
}

// ── msg ───────────────────────────────────────────────────────────────────────

func (b *Bot) handleMsg(chatID int64, args string) {
	if args == "" {
		b.reply(chatID, "Использование: /msg <текст>")
		return
	}
	if _, err := b.rcon.Execute("/say " + args); err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "✅ Сообщение отправлено")
}

// ── save ──────────────────────────────────────────────────────────────────────

func (b *Bot) handleSave(chatID int64) {
	if _, err := b.rcon.Execute("/server-save"); err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "💾 Сохранение выполнено")
}

// ── time ──────────────────────────────────────────────────────────────────────

func (b *Bot) handleTime(chatID int64) {
	resp, err := b.rcon.Execute("/time")
	if err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "⏱ "+resp)
}

// ── evolution ─────────────────────────────────────────────────────────────────

func (b *Bot) handleEvolution(chatID int64) {
	resp, err := b.rcon.Execute("/evolution")
	if err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "🦠 "+resp)
}

// ── restart (RCON /quit, Docker auto-restart) ─────────────────────────────────

func (b *Bot) handleRestart(chatID int64) {
	b.reply(chatID, "🔄 Перезапускаю сервер...")
	_, _ = b.rcon.Execute("/quit")
	time.Sleep(3 * time.Second)
	b.reply(chatID, "✅ Команда отправлена (Docker restart policy перезапустит контейнер)")
}

// ── stop container ────────────────────────────────────────────────────────────

func (b *Bot) handleStopServer(chatID int64) {
	b.reply(chatID, "⏹ Останавливаю контейнер...")
	if err := b.container.Stop(context.Background()); err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "✅ Контейнер остановлен")
}

// ── start container ───────────────────────────────────────────────────────────

func (b *Bot) handleStartServer(chatID int64) {
	b.reply(chatID, "▶️ Запускаю контейнер...")
	if err := b.container.Start(context.Background()); err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}
	b.reply(chatID, "✅ Контейнер запущен")
}

// ── getPassword ───────────────────────────────────────────────────────────────

func (b *Bot) handleGetPassword(chatID int64) {
	pw := b.passwords.Get()
	if pw == "" {
		b.reply(chatID, "❌ Пароль не сгенерирован")
		return
	}
	b.reply(chatID, "🔑 RCON пароль: `"+pw+"`")
}

// ── download save ─────────────────────────────────────────────────────────────

func (b *Bot) handleDownloadSave(chatID int64) {
	b.reply(chatID, "📦 Готовлю файл сохранения...")

	name, data, err := b.saves.LatestSave()
	if err != nil {
		b.reply(chatID, "❌ "+err.Error())
		return
	}

	b.replyDocument(chatID, name, data)
}

// ── upload save ───────────────────────────────────────────────────────────────

func (b *Bot) handleUploadSave(chatID int64, doc *tgbotapi.Document) {
	if !strings.HasSuffix(strings.ToLower(doc.FileName), ".zip") {
		b.reply(chatID, "❌ Ожидается .zip файл сохранения")
		return
	}

	b.reply(chatID, "📥 Загружаю файл...")

	data, err := b.downloadTelegramFile(doc.FileID)
	if err != nil {
		b.reply(chatID, "❌ Ошибка загрузки: "+err.Error())
		return
	}

	if err := b.saves.Replace(doc.FileName, data); err != nil {
		b.reply(chatID, "❌ Ошибка записи: "+err.Error())
		return
	}

	b.reply(chatID, fmt.Sprintf("✅ Сохранение «%s» загружено. Перезапусти сервер для применения.", doc.FileName))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (b *Bot) downloadTelegramFile(fileID string) ([]byte, error) {
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("getting file info: %w", err)
	}

	url := file.Link(b.api.Token)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	log.Printf("downloaded file %s: %d bytes", fileID, len(data))
	return data, nil
}
