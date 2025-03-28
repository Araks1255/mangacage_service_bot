package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) GetNewTeamsOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не модератор"))
		return
	}

	var teams []struct {
		ID          uint
		CreatedAt   time.Time
		Name        string
		Description string
		Creator     string
	}

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.name, t.description,
		users.user_name AS creator
		FROM teams_on_moderation AS t
		INNER JOIN users ON users.id = t.creator_id
		WHERE t.existing_id IS NULL`,
	).Scan(&teams)

	if len(teams) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет новых команд перевода на модерации"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(teams); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\n\nОтправлена на модерацию:\n%s",
			teams[i].ID, teams[i].Name, teams[i].Description, teams[i].Creator, teams[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тут будут инструкции по рассмотрению команды"))
}
