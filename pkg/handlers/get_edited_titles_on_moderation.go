package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (h handler) GetEditedTitlesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type editedTitle struct {
		gorm.Model
		Name        string
		Description string
		ExistingID  uint
		Creator     string
		Moder       string
		Author      string
		Genres      pq.StringArray `gorm:"type:[]TEXT"`
	}

	var titles []editedTitle
	h.DB.Raw(
		`SELECT t.id, t.created_at, t.existing_id, t.name, t.description,
		users.user_name AS creator, moders.user_name AS moder, authors.name AS author, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON users.id = t.creator_id
		INNER JOIN users AS moders ON moders.id = t.moderator_id
		LEFT JOIN authors ON authors.id = t.author_id`,
	).Scan(&titles)

	if len(titles) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет тайтлов, ожидающих подтверждения редактирования"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(titles); i++ {
		response = fmt.Sprintf(
			"id тайтла: %d\nid обращения: %d\n\nИзменения:\n Название: %s\n Описание: %s\n Автор: %s\n Жанры: %s\n\nВнёс изменения: %s\nПоследний редактировавший модератор: %s\n\nОтправлен на модерацию:\n%s",
			titles[i].ExistingID, titles[i].ID, titles[i].Name, titles[i].Description, titles[i].Author, strings.Join(titles[i].Genres, ", "), titles[i].Creator, titles[i].Moder, titles[i].CreatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы подробнее увидеть изменения тайтла (со старыми данными и обложкой), укажите id его обращения при вызове команды review_title\n\nПример: /review_title 2"))
}
