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

	desiredChapterID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди главы, которую хотите рассмотреть, после вызова функции\n\nПример: /review_chapter 2"))
		return
	}

	var existingChapterID uint
	h.DB.Raw("SELECT id FROM chapters WHERE id = ?", desiredChapterID).Scan(&existingChapterID)
	if existingChapterID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	var chapter struct {
		gorm.Model
		Name          string
		Description   string
		NumberOfPages int
		Volume        string
		Title         string
		Moder         string
	}

	h.DB.Raw(`SELECT chapters.id, chapters.created_at, chapters.updated_at, chapters.deleted_at, chapters.name, chapters.description,
		chapters.number_of_pages, volumes.name AS volume, titles.name AS title, users.user_name AS moder FROM chapters
		INNER JOIN volumes ON chapters.volume_id = volumes.id
		INNER JOIN titles ON volumes.title_id = titles.id
		LEFT JOIN users ON chapters.moderator_id = users.id
		WHERE chapters.on_moderation AND chapters.id = ?`,
		desiredChapterID,
	).Scan(&chapter)

	response := fmt.Sprintf(
		`ID главы: %d
		
		Глава тома %s тайтла %s
		
		Название: %s
		Описание: %s

		Количество страниц: %d

		Последний раз редактировавший модератор: %s
		
		Создана: %s
		Последний раз изменена: %s`,
		chapter.ID,
		chapter.Volume, chapter.Title,
		chapter.Name,
		chapter.Description,
		chapter.NumberOfPages,
		chapter.Moder,
		chapter.CreatedAt.Format(time.DateTime),
		chapter.UpdatedAt.Format(time.DateTime),
	)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	filter := bson.M{"chapter_id": chapter.ID}

	var result struct {
		Pages [][]byte `bson:"pages"`
	}

	for i := 0; i < chapter.NumberOfPages; i++ {
		projection := bson.M{"pages": bson.M{"$slice": []int{i, 1}}}

		err := h.Collection.FindOne(context.TODO(), filter, options.FindOne().SetProjection(projection)).Decode(&result)
		if err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
			return
		}

		h.Bot.Send(tgbotapi.NewMessage(tgUserID, fmt.Sprintf("Страница номер: %d", i+1)))
		h.Bot.Send(tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{"page", result.Pages[0]}))
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Всё"))
	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы одобрить главу, вызовите функцию /approve_chapter с указанием id главы\n\nЧтобы отвергнуть главу, вызовите функцию /reject_chapter с указанием id главы\n\nПримеры:\n/approve_chapter 12\n/reject_chapter 12"))
}
