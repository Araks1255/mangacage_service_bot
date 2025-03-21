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

	var existingChapterOnModerationID uint
	h.DB.Raw("SELECT id FROM chapters_on_moderation WHERE id = ?", desiredChapterOnModerationID).Scan(&existingChapterOnModerationID)
	if existingChapterOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	if result := h.DB.Exec("DELETE FROM chapters_on_moderation CASCADE WHERE id = ?", existingChapterOnModerationID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить главу"))
		return
	}

	filter := bson.M{"chapter_id": existingChapterOnModerationID}

	if _, err = h.ChaptersOnModerationPages.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить страницы главы (если глава ожидала подтверждения редактирования, то их и не было)"))
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию успешно отклонено"))
}
