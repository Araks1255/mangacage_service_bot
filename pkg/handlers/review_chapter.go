package handlers

import (
	"fmt"

	"github.com/Araks1255/mangacage/pkg/common/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ReviewChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredChapter := update.Message.CommandArguments()

	var chapter models.Chapter
	h.DB.Raw("SELECT * FROM chapters WHERE name = ?", desiredChapter).Scan(&chapter)
	if chapter.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	var chapterTitle string
	h.DB.Raw("SELECT name FROM titles WHERE id = ?", chapter.TitleID).Scan(&chapterTitle)

	createdAt := chapter.CreatedAt.Format("2006-01-02 15:04:05")

	response := fmt.Sprintf("Айди главы: %d\n\nНазвание: %s\nОписание: %s\nКоличество страниц: %d\nТайтл: %s\n\nСоздан в %s", chapter.ID, chapter.Name, chapter.Description, chapter.NumberOfPages, chapterTitle, createdAt)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Страницы главы:"))

	var path string
	for i := 0; i < chapter.NumberOfPages; i++ {
		path = fmt.Sprintf("%s/%d.jpg", chapter.Path, i)
		page := tgbotapi.NewPhoto(tgUserID, tgbotapi.FilePath(path))
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, fmt.Sprintf("Номер страницы: %d", i)))
		h.Bot.Send(page)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Всё"))
}
