package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

func (h handler) ReviewTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTitleOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди обращения тайтла, который хотите рассмотреть, после вызова функции\n\nПример: /review_title 2"))
		return
	}

	var existingTitleOnModerationID uint
	h.DB.Raw("SELECT id FROM titles_on_moderation WHERE id = ?", desiredTitleOnModerationID).Scan(&existingTitleOnModerationID)
	if existingTitleOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл не найден"))
		return
	}

	var titleCover TitleCover

	filter := bson.M{"title_id": existingTitleOnModerationID}

	if err := h.TitlesOnModerationCovers.FindOne(context.Background(), filter).Decode(&titleCover); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обложка не найдена"))
	}

	var title struct {
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

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.updated_at, t.deleted_at, t.name, t.description, t.existing_id,
		users.user_name AS creator, moders.user_name AS moder, authors.name AS author, teams.name AS team, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON t.creator_id = users.id
		LEFT JOIN users AS moders ON t.moderator_id = moders.id 
		INNER JOIN authors ON t.author_id = authors.id
		LEFT JOIN teams ON t.team_id = teams.id
		WHERE t.id = ?`, existingTitleOnModerationID,
	).Scan(&title)

	var response string

	if title.Moder == "" {
		response = fmt.Sprintf(
			"Причина обращения: создание\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nЖанры: %s\n\nОтправлен на модерацию:\n%s",
			title.ID, title.Name, title.Description, title.Creator, title.Author, strings.Join(title.Genres, ", "), title.CreatedAt.Format(time.DateTime),
		)
	} else {
		response = fmt.Sprintf(
			"id тайтла: %d\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nПоследний редактировавший модератор: %s\nЖанры: %s\n\nНа переводе у команды: %s\n\nОтправлен на модерацию:\n%s",
			title.ExistingID, title.ID, title.Name, title.Description, title.Creator, title.Author, title.Moder, strings.Join(title.Genres, ", "), title.Team, title.CreatedAt.Format(time.DateTime),
		)
	}

	cover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
		Name:  "cover",
		Bytes: titleCover.Cover,
	})
	cover.Caption = response

	h.Bot.Send(cover)
}
