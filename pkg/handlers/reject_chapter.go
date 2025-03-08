package handlers

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) RejectChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredChapter := strings.ToLower(update.Message.CommandArguments())
	if desiredChapter == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название главы, которую хотите отвергнуть, после команды\n\nПример: /reject_chapter Глава первая"))
		return
	}

	if result := h.DB.Exec("DELETE FROM chapters WHERE name = ?", desiredChapter); result.RowsAffected == 0 {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось отвергнуть главу"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава успешно отвергнута"))
}
