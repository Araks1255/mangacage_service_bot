package handlers

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ApproveVolume(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не админ и не модератор"))
		return
	}

	desiredVolumeID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id тома, который хотите одобрить, после вызова функции\n\nПример: /approve_volume 12"))
		return
	}

	var desiredVolumeName string
	h.DB.Raw("SELECT name FROM volumes WHERE id = ? AND on_moderation", desiredVolumeID).Scan(&desiredVolumeName)
	if desiredVolumeName == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том не найден"))
		return
	}

	if result := h.DB.Exec("UPDATE volumes SET on_moderation = false, moderator_id = ? WHERE id = ?", userID, desiredVolumeID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том успешно одобрен"))
}
