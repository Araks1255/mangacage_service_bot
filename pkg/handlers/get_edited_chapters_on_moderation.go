package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetEditedChaptersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, `Вы не являетесь модератором или администратором`))
		return
	}

	type editedChapter struct {
		gorm.Model
		Name          string
		Description   string
		ExistingID    uint
		NumberOfPages int
		Volume        string
		Title         string
		Creator       string
		Moder         string
	}

	var chapters []editedChapter
	h.DB.Raw(
		`SELECT c.id, c.created_at, c.updated_at, c.deleted_at, c.name, c.description, c.existing_id, c.number_of_pages,
		volumes.name AS volume, titles.name AS title, users.user_name AS creator, moders.user_name AS moder FROM chapters_on_moderation AS c
		INNER JOIN volumes ON volumes.id = c.volume_id
		INNER JOIN titles ON titles.id = volumes.title_id
		INNER JOIN users ON users.id = c.creator_id
		INNER JOIN users AS moders ON moders.id = c.moderator_id`,
	).Scan(&chapters)

	if len(chapters) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет глав, ожидающих подтверждения редактирования"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(chapters); i++ {
		response = fmt.Sprintf(
			"id главы: %d\nid обращения: %d\n\nГлава для тома %s тайтла %s\n\nНазвание: %s\nОписание: %s\nКоличество страниц: %d\nСоздатель: %s\nПоследний редактировавший модератор: %s\n\nОтправлена на моедрацию: %s",
			chapters[i].ExistingID, chapters[i].ID, chapters[i].Volume, chapters[i].Title, chapters[i].Name, chapters[i].Description, chapters[i].NumberOfPages, chapters[i].Creator, chapters[i].Moder, chapters[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы увидеть изменения, внесённые в главу, укажите id её обращения после вызова команды review_chapter_changes\n\nПример: /review_chapter_changes 2\n(Страницы главы изменяться не могут)"))
}
