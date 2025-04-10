package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

func (h handler) ReviewChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredChapterOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди обращения главы, которую хотите рассмотреть, после вызова функции\n\nПример: /review_chapter 2"))
		return
	}

	var existingChapterOnModerationID uint
	h.DB.Raw("SELECT id FROM chapters_on_moderation WHERE id = ?", desiredChapterOnModerationID).Scan(&existingChapterOnModerationID)
	if existingChapterOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	var chapterOnModeration struct {
		gorm.Model
		Name          string
		Description   string
		NumberOfPages int
		ExistingID    uint
		Volume        string
		Title         string
		Creator       string
		Moder         string
	}

	h.DB.Raw(
		`SELECT c.id, c.created_at, c.name, c.description, c.number_of_pages, c.existing_id,
		volumes.name AS volume, titles.name AS title, users.user_name AS creator, moders.user_name AS moder
		FROM chapters_on_moderation AS c
		INNER JOIN volumes ON c.volume_id = volumes.id
		INNER JOIN titles ON volumes.title_id = titles.id
		INNER JOIN users ON users.id = c.creator_id
		LEFT JOIN users AS moders ON moders.id = c.moderator_id
		WHERE c.id = ?`, desiredChapterOnModerationID,
	).Scan(&chapterOnModeration)

	isChapterNew := chapterOnModeration.ExistingID == 0

	var response string

	if !isChapterNew {
		response = fmt.Sprintf(
			"Причина обращения: редактирование\nid главы: %d\nid обращения: %d\nГлава для тома %s тайтла %s\n\nИзменения:\n Название: %s\n Описание: %s\n\nВнёс изменения: %s\nПоследний редактировавший модератор: %s\n\nОтпралена на модерацию:\n%s",
			chapterOnModeration.ExistingID, chapterOnModeration.ID, chapterOnModeration.Volume, chapterOnModeration.Title, chapterOnModeration.Name, chapterOnModeration.Description, chapterOnModeration.Creator, chapterOnModeration.Moder, chapterOnModeration.CreatedAt.Format(time.DateTime),
		)

		h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

		var chapter struct {
			Name        string
			Description string
		}

		h.DB.Raw("SELECT name, description FROM chapters WHERE id = ?", chapterOnModeration.ExistingID).Scan(&chapter)

		response = "Изменения:\n\n"

		if chapterOnModeration.Name != "" {
			response += fmt.Sprintf("Название с \"%s\" на \"%s\"\n", chapter.Name, chapterOnModeration.Name)
		}

		if chapterOnModeration.Description != "" {
			response += fmt.Sprintf("Описание с \"%s\" на \"%s\"", chapter.Description, chapterOnModeration.Description)
		}

		h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))
		return
	}

	response = fmt.Sprintf(
		"Причина обращения: создание\nid обращения: %d\n\nГлава для тома %s тайтла %s\n\nНазвание: %s\nОписание: %s\nКоличество страниц: %d\nСоздатель: %s\n\nОтправлена на модерацию:\n%s",
		chapterOnModeration.ID, chapterOnModeration.Volume, chapterOnModeration.Title, chapterOnModeration.Name, chapterOnModeration.Description, chapterOnModeration.NumberOfPages, chapterOnModeration.Creator, chapterOnModeration.CreatedAt.Format(time.DateTime),
	)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	filter := bson.M{"chapter_id": chapterOnModeration.ID}

	var result struct {
		Pages [][]byte `bson:"pages"`
	}

	var page tgbotapi.PhotoConfig
	for i := 0; i < chapterOnModeration.NumberOfPages; i++ {
		projection := bson.M{"pages": bson.M{"$slice": []int{i, 1}}}

		err := h.ChaptersOnModerationPages.FindOne(context.TODO(), filter, options.FindOne().SetProjection(projection)).Decode(&result)
		if err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
			return
		}

		page = tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
			Name:  "page",
			Bytes: result.Pages[0],
		})
		page.Caption = fmt.Sprintf("Страница номер %d", i+1)

		h.Bot.Send(page)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы одобрить главу, вызовите функцию /approve_chapter с указанием id её обращения\n\nЧтобы отвергнуть главу, вызовите функцию /reject_chapter с указанием id её обращения\n\nПримеры:\n/approve_chapter 12\n/reject_chapter 12"))
}
