package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetEditedVolumesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type editedVolume struct {
		gorm.Model
		Name        string
		Description string
		ExistingID  uint
		Title       string
		Creator     string
		Moder       string
	}

	var volumes []editedVolume
	h.DB.Raw(
		`SELECT v.id, v.created_at, v.name, v.description, v.existing_id,
		titles.name AS title, users.user_name AS creator, moders.user_name AS moder FROM volumes_on_moderation AS v
		INNER JOIN titles ON titles.id = v.title_id
		INNER JOIN users ON users.id = v.creator_id
		LEFT JOIN users AS moders ON moders.id = v.moderator_id`,
	).Scan(&volumes)

	if len(volumes) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет томов, ожидающих подтверждения редактирования"))
		return
	}

	var (
		msg      tgbotapi.MessageConfig
		response string
	)

	for i := 0; i < len(volumes); i++ {
		response = fmt.Sprintf(
			"id тома: %d\nid обращения: %d\nТайтл: %s\n\nИзменения:\n Название: %s\n Описание: %s\n\nВнёс изменения: %s\nПоследний редактировавший модератор: %s\n\nОтправлен на модерацию:\n%s",
			volumes[i].ExistingID, volumes[i].ID, volumes[i].Title, volumes[i].Name, volumes[i].Description, volumes[i].Creator, volumes[i].Moder, volumes[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы подробнее рассмотреть изменения в томе, укажите его id обращения после вызова команды review_volume\n\nПример: /review_volume 2"))
}
