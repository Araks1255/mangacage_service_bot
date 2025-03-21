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

	var titleOnModeration struct {
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
	).Scan(&titleOnModeration)

	isTitleNew := titleOnModeration.Moder == ""

	var response string

	if isTitleNew {
		response = fmt.Sprintf(
			"Причина обращения: создание\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nЖанры: %s\n\nОтправлен на модерацию:\n%s",
			titleOnModeration.ID, titleOnModeration.Name, titleOnModeration.Description, titleOnModeration.Creator, titleOnModeration.Author, strings.Join(titleOnModeration.Genres, ", "), titleOnModeration.CreatedAt.Format(time.DateTime),
		)

		var titleCover TitleCover

		filter := bson.M{"title_id": existingTitleOnModerationID}

		if err := h.TitlesOnModerationCovers.FindOne(context.Background(), filter).Decode(&titleCover); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обложка не найдена"))
		}

		cover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
			Name:  "cover",
			Bytes: titleCover.Cover,
		})
		cover.Caption = response

		h.Bot.Send(cover)
		return
	}

	response = fmt.Sprintf(
		"Причина обращения: редактирование\nid тайтла: %d\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nПоследний редактировавший модератор: %s\nЖанры: %s\n\nНа переводе у команды: %s\n\nОтправлен на модерацию:\n%s",
		titleOnModeration.ExistingID, titleOnModeration.ID, titleOnModeration.Name, titleOnModeration.Description, titleOnModeration.Creator, titleOnModeration.Author, titleOnModeration.Moder, strings.Join(titleOnModeration.Genres, ", "), titleOnModeration.Team, titleOnModeration.CreatedAt.Format(time.DateTime),
	)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	var title struct {
		Name        string
		Description string
		Author      string
		Genres      pq.StringArray `gorm:"type:text[]"`
	}

	h.DB.Raw(
		`SELECT t.name, t.description, authors.name AS author,
		(
		SELECT ARRAY(
		SELECT genres.name FROM genres
		INNER JOIN title_genres ON genres.id = title_genres.genre_id
		INNER JOIN titles ON title_genres.title_id = titles.id
		WHERE titles.id = t.id) AS genres
		)
		FROM titles AS t
		INNER JOIN authors ON authors.id = t.author_id
		INNER JOIN teams ON teams.id = t.team_id
		WHERE NOT t.on_moderation
		AND t.id = ?`, titleOnModeration.ExistingID).Scan(&title)

	response = "Изменения:\n\n"

	if title.Name != titleOnModeration.Name {
		response += fmt.Sprintf("Название с %s на %s\n", title.Name, titleOnModeration.Name)
	}
	if title.Description != titleOnModeration.Description {
		response += fmt.Sprintf("Описание с %s на %s\n", title.Description, titleOnModeration.Description)
	}
	if title.Author != titleOnModeration.Author {
		response += fmt.Sprintf("Автор с %s на %s\n", title.Author, titleOnModeration.Author)
	}

	if len(titleOnModeration.Genres) != 0 { // По задумке, если изменений нет, то будет ничего писаться, следовательно длина этого слайса будет 0
		response += fmt.Sprintf("Жанры с %s на %s\n", strings.Join(title.Genres, ", "), strings.Join(titleOnModeration.Genres, ", "))
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	var newTitleCover TitleCover

	filter := bson.M{"title_id": titleOnModeration.ID}

	if err := h.TitlesOnModerationCovers.FindOne(context.Background(), filter).Decode(&newTitleCover); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обложка не обновлена"))
		return
	}

	var oldTitleCover TitleCover

	filter = bson.M{"title_id": titleOnModeration.ExistingID}

	if err = h.TitlesCovers.FindOne(context.Background(), filter).Decode(&oldTitleCover); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Старая обложка не найдена"))
	} else {
		oldCover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
			Name:  "cover",
			Bytes: oldTitleCover.Cover,
		})
		oldCover.Caption = "Старая обложка"
		h.Bot.Send(oldCover)
	}

	newCover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
		Name:  "cover",
		Bytes: newTitleCover.Cover,
	})
	newCover.Caption = "Новая обложка"

	h.Bot.Send(newCover)
}
