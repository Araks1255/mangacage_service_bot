package handlers

import (
	"fmt"

	"github.com/Araks1255/mangacage/pkg/common/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) GetTitlesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsersTgIds[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	var titles []models.Title
	h.DB.Raw("SELECT * FROM titles WHERE on_moderation").Scan(&titles)

	if len(titles) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет тайтлов на модерации"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, fmt.Sprintf("Тайтлов на модерации: %d", len(titles))))

	var creatorName, authorName, createdAt string

	for i := 0; i < len(titles); i++ {
		h.DB.Raw("SELECT user_name FROM users WHERE id = ?", titles[i].CreatorID).Scan(&creatorName)
		h.DB.Raw("SELECT name FROM authors WHERE id = ?", titles[i].AuthorID).Scan(&authorName)

		createdAt = titles[i].CreatedAt.Format("2006-01-02 15:04:05")

		response := fmt.Sprintf("ID тайтла: %d\n\nНазвание: %s\nОписание: %s\nИмя создателя: %s\nИмя автора: %s\n\nОтправлен на модерацию в %s",
			titles[i].ID,
			titles[i].Name,
			titles[i].Description,
			creatorName,
			authorName,
			createdAt)

		msg := tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}
}
