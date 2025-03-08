package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetTitlesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type result struct {
		gorm.Model
		Name        string
		Description string
		CreatorName string
		ModerName   string
		AuthorName  string
	}

	var titles []result
	h.DB.Raw(`SELECT titles.id, titles.created_at, titles.updated_at, titles.deleted_at, titles.name, titles.description,
	users.user_name AS creator_name, moders.user_name AS moder_name, authors.name AS author_name FROM titles
	INNER JOIN users ON titles.creator_id = users.id
	INNER JOIN authors ON titles.author_id = authors.id
	LEFT JOIN users AS moders ON titles.moderator_id = moders.id
	WHERE titles.on_moderation`).Scan(&titles)

	if len(titles) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет тайтлов на модерации"))
		return
	}

	var (
		titleGenres []string
		msg         tgbotapi.MessageConfig
		response    string
	)

	for i := 0; i < len(titles); i++ {
		h.DB.Raw(`SELECT genres.name FROM genres
		INNER JOIN title_genres ON genres.id = title_genres.genre_id
		INNER JOIN titles ON title_genres.title_id = titles.id
		WHERE titles.id = ?`, titles[i].ID,
		).Scan(&titleGenres)

		response = fmt.Sprintf(
			`id тайтла: %d

			Название: %s
			Описание: %s
			Создатель: %s
			Автор: %s

			Жанры: %s
			
			Последний редактировавший модератор: %s
			
			Создан: %s
			Последний раз изменён: %s`,
			titles[i].ID,
			titles[i].Name,
			titles[i].Description,
			titles[i].CreatorName,
			titles[i].AuthorName,
			strings.Join(titleGenres, ", "),
			titles[i].ModerName,
			titles[i].CreatedAt.Format(time.DateTime),
			titles[i].UpdatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Всё"))
}
