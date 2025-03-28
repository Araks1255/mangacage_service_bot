package handlers

import (
	"context"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (h handler) RejectChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredChapterOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения главы, которое хотите отклонить\n\n Пример: /reject_chapter 2"))
		return
	}

	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	defer tx.Rollback()

	var ChapterOnModerationID, chapterID uint
	row := h.DB.Raw("SELECT id, existing_id FROM chapters_on_moderation WHERE id = ?", desiredChapterOnModerationID).Row()

	if err := row.Scan(&ChapterOnModerationID, chapterID); err != nil {
		log.Println(err)
	}

	if ChapterOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	doesChapterExist := chapterID != 0

	if result := h.DB.Exec("DELETE FROM chapters_on_moderation WHERE id = ?", ChapterOnModerationID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить главу"))
		return
	}

	if !doesChapterExist {
		filter := bson.M{"chapter_id": ChapterOnModerationID}
		if _, err = h.ChaptersOnModerationPages.DeleteOne(context.TODO(), filter); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить страницы главы"))
			return
		}
	}

	tx.Commit()

	if !doesChapterExist {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию новой главы успешно отклонено"))
	} else {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию для редактирования главы успешно отклонено"))
	}
}
