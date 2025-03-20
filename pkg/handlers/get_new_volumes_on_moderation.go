package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetNewVolumesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type volume struct {
		gorm.Model
		Name        string
		Description string
		Title       string
		Creator     string
	}

	var volumes []volume
	h.DB.Raw(`SELECT v.id, v.created_at, v.updated_at, v.deleted_at, v.name, v.description,
		titles.name AS title, users.user_name AS creator FROM volumes_on_moderation as v
		INNER JOIN titles ON v.title_id = titles.id
		INNER JOIN users ON v.creator_id = users.id`,
	).Scan(&volumes)

	if len(volumes) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет томов на модерации"))
		return
	}

	var (
		msg      tgbotapi.MessageConfig
		response string
	)

	for i := 0; i < len(volumes); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\n\nТом для тайтла %s\nНазвание: %s\nОписание: %s\nСоздатель: %s\n\nОтправлен на модерацию:\n%s",
			volumes[i].ID, volumes[i].Title, volumes[i].Name, volumes[i].Description, volumes[i].Creator, volumes[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Всё (все данные томов выведены, больше нет)"))
}
