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

	desiredVolumeOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения тома, которое хотите отклонить\n\n Пример: /reject_volume 2"))
		return
	}

	var existingVolumeOnModerationID uint
	h.DB.Raw("SELECT id FROM volumes_on_moderation WHERE id = ?", desiredVolumeOnModerationID).Scan(&existingVolumeOnModerationID)
	if existingVolumeOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том не найден"))
		return
	}

	if result := h.DB.Exec("DELETE FROM volumes_on_moderation CASCADE WHERE id = ?", existingVolumeOnModerationID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить том"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию успешно отклонено"))
}
