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
		Team        string
		Genres      pq.StringArray `gorm:"type:[]TEXT"`
	}

	var titles []editedTitle
	h.DB.Raw(
		`SELECT t.id, t.created_at, t.updated_at, t.deleted_at, t.name, t.description, t.existing_id,
		users.user_name AS creator, moders.user_name AS moder, authors.name AS author, teams.name AS team, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON t.creator_id = users.id
		INNER JOIN users AS moders ON t.moderator_id = moders.id 
		INNER JOIN authors ON t.author_id = authors.id
		LEFT JOIN teams ON t.team_id = teams.id`, // В строке 5 идет inner join к модераторам, а это значит, что берутся только те записи, в которых есть moderator_id, то есть, которые уже прошли изначальную модерацию, а сейчас пришли на редактирование
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
			"id тайтла: %d\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nПоследний редактировавший модератор: %s\nЖанры: %s\n\nНа переводе у команды: %s\n\nОтправлен на модерацию:\n%s",
			titles[i].ExistingID, titles[i].ID, titles[i].Name, titles[i].Description, titles[i].Creator, titles[i].Author, titles[i].Moder, strings.Join(titles[i].Genres, ", "), titles[i].Team, titles[i].CreatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы просмотреть изменения в конкретном тайтле, укажите название этого тайтла после вызова команды review_title_changes\n\nПример: /review_title_changes Мёртвый аккаунт"))
}
