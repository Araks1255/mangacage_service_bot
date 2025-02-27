package handlers

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ApproveTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTitleID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите ID тайтла, который хотите одобрить, через пробел после команды\n\nПример: /approve_title 1"))
		return
	}

	if result := h.DB.Exec("UPDATE titles SET on_moderation = false, moderator_id = ? WHERE id = ?", userID, desiredTitleID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось снять тайтл с модерации. Возможно вы ошиблись в айди"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно снят с модерации"))
}
