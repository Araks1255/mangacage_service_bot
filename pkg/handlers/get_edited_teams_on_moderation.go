package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) GetEditedTeamsOnModeration(update tgbotapi.Update) {
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
		Existing    string
		Creator     string
	}

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.name, t.description,
		teams.name AS existing, users.user_name AS creator
		FROM teams_on_moderation AS t
		INNER JOIN teams ON teams.id = t.existing_id
		INNER JOIN users ON users.id = t.creator_id`,
	).Scan(&teams)

	if len(teams) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет команд перевода, ожидающих подтверждения редактирования"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(teams); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\n\nИзменения для команды %s:\n Название: %s\n Описание: %s\n\nОтправил на модерацию: %s\n\nОтправлено на модерацию:\n%s",
			teams[i].ID, teams[i].Existing, teams[i].Name, teams[i].Description, teams[i].Creator, teams[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тут будут инструкции по рассмотрению команды"))
}
