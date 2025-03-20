package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetNewChaptersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, `Вы не являетесь модератором или администратором`))
		return
	}

	type newChapter struct {
		gorm.Model
		Name          string
		Description   string
		NumberOfPages int
		Volume        string
		Title         string
		Creator       string
	}

	var chapters []newChapter
	h.DB.Raw(
		`SELECT c.id, c.created_at, c.updated_at, c.deleted_at, c.name, c.description, c.number_of_pages,
		volumes.name AS volume, titles.name AS title, users.user_name AS creator FROM chapters_on_moderation AS c
		INNER JOIN volumes ON volumes.id = c.volume_id
		INNER JOIN titles ON titles.id = volumes.title_id
		INNER JOIN users ON users.id = c.creator_id
		WHERE c.moderator_id IS NULL`,
	).Scan(&chapters)

	if len(chapters) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, `Нет новых глав на модерации`))
		return
	}

	var (
		msg      tgbotapi.MessageConfig
		response string
	)

	for i := 0; i < len(chapters); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\nГлава для тома %s тайтла %s\n\nНазвание: %s\nОписание: %s\nКоличество страниц: %d\nСоздатель: %s\n\nОтправлена на модерацию: %s",
			chapters[i].ID, chapters[i].Volume, chapters[i].Title, chapters[i].Name, chapters[i].Description, chapters[i].NumberOfPages, chapters[i].Creator, chapters[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы увидеть страницы главы, укажите id её обращения при вызове функции review_chapter\n\nПример: /review_chapter 2"))
}
