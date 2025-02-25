package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) GetChaptersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsersTgIds[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	var chapters []string
	h.DB.Raw("SELECT name FROM chapters WHERE on_moderation").Scan(&chapters)

	if len(chapters) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет глав на модерации"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, fmt.Sprintf("Количество глав на модерации: %d", len(chapters))))
	for i := 0; i < len(chapters); i++ {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, chapters[i]))
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Для рассмотрения главы вызовите функцию /review_chapter с указанием названия необходимой главы\n\nПример: /review_chapter Мёртвый аккаунт"))
}
