package handlers

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) RejectVolume(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь админом или модератором"))
		return
	}

	desiredVolumeID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id тома, который хотите отвергнуть\n\nПример: /reject_volume 12"))
		return
	}

	var desiredVolumeName string
	h.DB.Raw("SELECT name FROM volumes WHERE id = ? AND on_moderation", desiredVolumeID).Scan(&desiredVolumeName)
	if desiredVolumeName == "" { // Аналогично, на случай, если придётся отправлять уведомления
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том не найден"))
		return
	}

	if result := h.DB.Exec("DELETE FROM volumes WHERE id = ?", desiredVolumeID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том успешно отвергнут"))
}
