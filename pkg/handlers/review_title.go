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
		Genres      pq.StringArray `gorm:"type:[]TEXT"`
	}

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.existing_id, t.name, t.description,
		users.user_name AS creator, moders.user_name AS moder, authors.name AS author, t.genres
		FROM titles_on_moderation AS t
		INNER JOIN users ON users.id = t.creator_id
		INNER JOIN users AS moders ON moders.id = t.moderator_id
		LEFT JOIN authors ON authors.id = t.author_id
		WHERE t.id = ?`, existingTitleOnModerationID,
	).Scan(&titleOnModeration)

	isTitleNew := titleOnModeration.ExistingID == 0

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
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы одобрить тайтл, вызовите команду approve_title с указанием id его обращения\nЧтобы отклонить обращение на модерацию тайтла, вызовите команду reject_title с указанием id его обращения\n\nПримеры:\n/approve_title 2\n/reject_title 2"))
		return
	}

	var title struct {
		gorm.Model
		Name        string
		Description string
		Author      string
		Creator     string
		Moder       string
		Team        string
		Genres      pq.StringArray `gorm:"type:text[]"`
	}

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.name, t.description, authors.name AS author, users.user_name AS creator, moders.user_name AS moder, teams.name AS team,
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
		INNER JOIN users ON users.id = t.creator_id
		INNER JOIN users AS moders ON moders.id = t.moderator_id
		WHERE t.id = ?`, titleOnModeration.ExistingID).Scan(&title)

	response = fmt.Sprintf(
		"Причина обращения: редактирование\n\nИнформация о тайтле (на данный момент):\nid: %d\nНазвание: %s\nОписание: %s\nСоздатель: %s\nАвтор: %s\nПоследний редактировавший модератор: %s\nЖанры: %s\n\nНа переводе у команды: %s\n\nСоздан:\n%s",
		title.ID, title.Name, title.Description, title.Creator, title.Author, title.Moder, strings.Join(title.Genres, ", "), title.Team, title.CreatedAt.Format(time.DateTime),
	)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	response = "Изменения:\n\n"

	if titleOnModeration.Name != "" {
		response += fmt.Sprintf("Название с \"%s\" на \"%s\"\n", title.Name, titleOnModeration.Name)
	}
	if titleOnModeration.Description != "" {
		response += fmt.Sprintf("Описание с \"%s\" на \"%s\"\n", title.Description, titleOnModeration.Description)
	}
	if titleOnModeration.Author != "" {
		response += fmt.Sprintf("Автор с \"%s\" на \"%s\"\n", title.Author, titleOnModeration.Author)
	}

	if len(titleOnModeration.Genres) != 0 { // По задумке, если изменений нет, то будет ничего писаться, следовательно длина этого слайса будет 0
		response += fmt.Sprintf("Жанры с \"%s\" на \"%s\"\n", strings.Join(title.Genres, ", "), strings.Join(titleOnModeration.Genres, ", "))
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

	if err = h.TitlesCovers.FindOne(context.TODO(), filter).Decode(&oldTitleCover); err != nil {
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

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы одобрить тайтл, выз"))
}
