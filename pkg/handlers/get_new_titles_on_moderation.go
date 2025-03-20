package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (h handler) GetNewTitlesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type newTitle struct {
		gorm.Model
		Name        string
		Description string
		Creator     string
		Author      string
		Genres      pq.StringArray `gorm:"type:[]TEXT"`
	}

	var titles []newTitle
	h.DB.Raw(
		`SELECT t.id, t.created_at, t.updated_at, t.deleted_at, t.name, t.description,
		users.user_name AS creator, authors.name AS author, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON t.creator_id = users.id
		INNER JOIN authors ON t.author_id = authors.id
		WHERE t.moderator_id IS NULL`,
	).Scan(&titles)

	if len(titles) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет тайтлов на модерации"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(titles); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nЖанры: %s\n\nОтправлен на модерацию:\n%s",
			titles[i].ID,
			titles[i].Name,
			titles[i].Description,
			titles[i].Creator,
			titles[i].Author,
			strings.Join([]string(titles[i].Genres), ", "),
			titles[i].CreatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы увидеть все данные конкретного тайтла (включая обложку), напишите название тайтла после вызова команды review_title\n\nПример: /review_title Мёртвый аккаунт"))
}
