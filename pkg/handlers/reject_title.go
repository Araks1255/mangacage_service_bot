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

	desiredTitleName := strings.ToLower(update.Message.CommandArguments())
	if desiredTitleName == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название тайтла, который хотите не принять, после вызова команды\n\nПример: /reject_title Мёртвый аккаунт"))
		return
	}

	var existingTitleID uint
	h.DB.Raw("SELECT id FROM titles WHERE name = ? AND on_moderation", desiredTitleName).Scan(&existingTitleID)
	if existingTitleID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл не найден. Введите название тайтла, который хотите отвергнкть, через пробел после вызова функции\n\nПример: /reject_title Мертвый аккаунт"))
		return
	}

	if result := h.DB.Exec("DELETE FROM titles WHERE id = ? CASCADE", existingTitleID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно отвергнут"))
}
