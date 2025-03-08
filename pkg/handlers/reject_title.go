package handlers

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) RejectTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTitle := strings.ToLower(update.Message.CommandArguments())
	if desiredTitle == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название тайтла, который хотите не принять, после вызова команды\n\nПример: /reject_title Мёртвый аккаунт"))
		return
	}

	if result := h.DB.Exec("DELETE FROM titles WHERE name = ?", desiredTitle); result.RowsAffected == 0 {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось не принять тайтл"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно отвергнут"))
}
