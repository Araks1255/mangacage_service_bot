package handlers

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func (h handler) GetTitlesOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	type title struct {
		gorm.Model
		Name        string
		Description string
		ExistingID  uint
		Creator     string
		Moder       string
		Author      string
		Team        string
		Genres      pq.StringArray `gorm:"type:[]TEXT"`
	}

	var titles []title
	h.DB.Raw(`SELECT t.id, t.created_at, t.updated_at, t.deleted_at, t.name, t.description, t.existing_id,
		users.user_name AS creator, moders.user_name AS moder, authors.name AS author, teams.name AS team, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON t.creator_id = users.id
		LEFT JOIN users AS moders ON t.moderator_id = moders.id
		INNER JOIN authors ON t.author_id = authors.id
		LEFT JOIN teams ON t.team_id = teams.id`).Scan(&titles)

	if len(titles) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет тайтлов на модерации"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(titles); i++ {
		switch titles[i].Moder {
		case "":
			response = fmt.Sprintf(
				"id обращения: %d\nПричина обращения: создание\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nЖанры: %s\n\nОтправлен на модерацию:\n%s",
				titles[i].ID,
				titles[i].Name,
				titles[i].Description,
				titles[i].Creator,
				titles[i].Author,
				strings.Join([]string(titles[i].Genres), ", "),
				titles[i].CreatedAt.Format(time.DateTime),
			)
		default:
			response = fmt.Sprintf(
				"id тайтла (существующего): %d\nПричина обращения: редактирование\n\nid обращения: %d\nНазвание: %s\nОписание: %s\nАвтор: %s\nЖанры:%s\nСоздатель: %s\nПоследний редактировавший модератор: %s\nНа переводе у команды %s\n\nОтправлен на модерацию:\n%s",
				titles[i].ExistingID,
				titles[i].ID,
				titles[i].Name,
				titles[i].Description,
				titles[i].Author,
				strings.Join(titles[i].Genres, ", "),
				titles[i].Creator,
				titles[i].Moder,
				titles[i].Team,
				titles[i].CreatedAt.Format(time.DateTime),
			)
		}

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы увидеть все данные конкретного тайтла (включая обложку), напишите название тайтла после вызова команды review_title\n\nПример: /review_title Мёртвый аккаунт"))
}
