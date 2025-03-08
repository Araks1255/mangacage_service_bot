package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetVolumesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type result struct {
		gorm.Model
		Name        string
		Description string
		Title       string
		Creator     string
		Moder       string
	}

	var volumes []result
	h.DB.Raw(`SELECT volumes.id, volumes.created_at, volumes.updated_at, volumes.deleted_at, volumes.name, volumes.description,
		titles.name AS title, users.user_name AS creator, moders.user_name AS moder FROM volumes
		INNER JOIN titles ON volumes.title_id = titles.id
		INNER JOIN users ON volumes.creator_id = users.id
		LEFT JOIN users AS moders ON volumes.moderator_id = moders.id
		WHERE volumes.on_moderation`,
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
		response = fmt.Sprintf(`
		ID тома: %d
		
		Том для тайтла: %s

		Название: %s
		Описание: %s
		Создатель %s
		
		Последний раз редактировавший модератор: %s
		
		Создан: %s
		Последний раз изменён: %s`,
			volumes[i].ID,
			volumes[i].Title,
			volumes[i].Name,
			volumes[i].Description,
			volumes[i].Creator,
			volumes[i].Moder,
			volumes[i].CreatedAt.Format(time.DateTime),
			volumes[i].UpdatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Всё"))
}
