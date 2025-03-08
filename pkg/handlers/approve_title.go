package handlers

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ApproveTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTitle := strings.ToLower(update.Message.CommandArguments())
	if desiredTitle == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название тайтла, который хотите одобрить, после вызова команды\n\nПример: /approve_title Мёртвый аккаунт"))
		return
	}

	if result := h.DB.Exec("UPDATE titles SET on_moderation = false, moderator_id = ? WHERE name = ?", userID, desiredTitle); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось снять тайтл с модерации. Возможно вы ошиблись в айди"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно снят с модерации"))
}
