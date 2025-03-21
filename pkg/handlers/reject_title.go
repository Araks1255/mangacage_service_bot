package handlers

import (
	"log"
	"strconv"
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (h handler) RejectTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTitleOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения тайтла, которое хотите отклонить\n\n Пример: /reject_title 2"))
		return
	}

	var existingTitleOnModerationID uint
	h.DB.Raw("SELECT id FROM titles_on_moderation WHERE id = ?", desiredTitleOnModerationID).Scan(&existingTitleOnModerationID)
	if existingTitleOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл не найден"))
		return
	}

	if result := h.DB.Exec("DELETE FROM titles_on_moderation CASCADE WHERE id = ?", existingTitleOnModerationID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить тайтл"))
		return
	}

	filter := bson.M{"title_id":existingTitleOnModerationID}

	if _, err = h.TitlesOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить обложку тайтла (если тайтл ожидал редактирования и обложка не была изменена, то её и не было)"))
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию успешно отклонено"))
}
