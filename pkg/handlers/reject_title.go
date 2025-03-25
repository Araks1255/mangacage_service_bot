package handlers

import (
	"context"
	"database/sql"
	"log"
	"strconv"

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

	tx := h.DB.Begin()
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
	defer tx.Rollback()

	var titleID, titleOnModerationID sql.NullInt64

	row := tx.Raw("SELECT existing_id, id FROM titles_on_moderation WHERE id = ?", desiredTitleOnModerationID).Row()

	if err := row.Scan(&titleID, &titleOnModerationID); err != nil {
		log.Println(err)
	}

	if !titleID.Valid && !titleOnModerationID.Valid {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл не найден"))
		return
	}

	if result := tx.Exec("DELETE FROM titles_on_moderation WHERE id = ?", titleOnModerationID.Int64); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить тайтл"))
		return
	}

	var filter bson.M
	if titleID.Valid {
		filter = bson.M{"title_id": titleID.Int64}
	} else {
		filter = bson.M{"title_on_moderation_id": titleOnModerationID.Int64}
	}

	if _, err = h.TitlesOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить обложку тайтла (если тайтл ожидал редактирования и обложка не была изменена, то её и не было)"))
	}

	tx.Commit()

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию успешно отклонено"))
}
