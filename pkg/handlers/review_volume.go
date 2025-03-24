package handlers

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) ReviewVolume(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredVolumeOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди обращения тома, который хотите рассмотреть, после вызова функции\n\nПример: /review_volume 2"))
		return
	}

	var existingVolumeOnModerationID uint
	h.DB.Raw("SELECT id FROM volumes_on_moderation WHERE id = ?", desiredVolumeOnModerationID).Scan(&existingVolumeOnModerationID)
	if existingVolumeOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том не найден"))
		return
	}

	var volumeOnModeration struct {
		gorm.Model
		Name        string
		Description string
		ExistingID  uint
		Creator     string
		Moder       string
		Title       string
	}

	h.DB.Raw(
		`SELECT v.id, v.created_at, v.name, v.description, v.existing_id,
		users.user_name AS creator, moders.user_name AS moder, titles.name AS title
		FROM volumes_on_moderation AS v
		INNER JOIN users ON users.id = v.creator_id
		LEFT JOIN users AS moders ON moders.id = v.moderator_id
		INNER JOIN titles ON titles.id = v.title_id
		WHERE v.id = ?`, existingVolumeOnModerationID,
	).Scan(&volumeOnModeration)

	isVolumeNew := volumeOnModeration.ExistingID == 0

	var response string

	if isVolumeNew {
		response = fmt.Sprintf(
			"Причина обращения: создание\nid образения: %d\n\nНазвание: %s\nОписание: %s\nТайтл: %s\n\nСоздатель: %s\nПоследний редактировавший модератор: %s\n\nОтправлен на модерацию:\n%s",
			volumeOnModeration.ID, volumeOnModeration.Name, volumeOnModeration.Description, volumeOnModeration.Title, volumeOnModeration.Creator, volumeOnModeration.Moder, volumeOnModeration.CreatedAt.Format(time.DateTime),
		)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "тут будет инструкция по одобрению и неодобрению"))
		return
	}

	var volume struct { // Потом в отдельные структуры вынесу, пока хардкодом
		gorm.Model
		Name        string
		Description string
		Creator     string
		Moder       string
		Title       string
	}

	h.DB.Raw(
		`SELECT v.id, v.created_at, v.name, v.description,
		users.user_name AS creator, moders.user_name AS moder, titles.name AS title
		FROM volumes AS v
		INNER JOIN users ON users.id = v.creator_id
		LEFT JOIN users AS moders ON moders.id = v.moderator_id
		INNER JOIN titles ON titles.id = v.title_id
		WHERE v.id = ?`, volumeOnModeration.ExistingID,
	).Scan(&volume) // пока сырые запросы делаю, потом возможно на прелоады перейду

	response = fmt.Sprintf(
		"Причина обращения: редактирование\n\nИнформация о томе (на данный момент):\nid: %d\nНазвание: %s\nОписание: %s\nТайтл: %s\n\nСоздатель: %s\nПоследний редактировавший модератор: %s\n\nСоздан:\n%s",
		volume.ID, volume.Name, volume.Description, volume.Title, volume.Creator, volume.Moder, volume.CreatedAt.Format(time.DateTime),
	)
	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	response = "Изменения:\n\n"

	if volumeOnModeration.Name != "" {
		response += fmt.Sprintf("Название с \"%s\" на \"%s\"\n", volume.Name, volumeOnModeration.Name)
	}
	if volumeOnModeration.Description != "" {
		response += fmt.Sprintf("Описание с \"%s\" на \"%s\"\n", volume.Description, volumeOnModeration.Description)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))
}
