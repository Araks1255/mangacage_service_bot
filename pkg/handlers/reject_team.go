package handlers

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (h handler) RejectTeam(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTeamOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "*Инструкции по правильному использованию команды*"))
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

	var (
		teamID             sql.NullInt64
		teamOnModerationID uint
	)

	row := tx.Raw("SELECT existing_id, id FROM teams_on_moderation WHERE id = ?", desiredTeamOnModerationID).Row()

	if err := row.Scan(&teamID, &teamOnModerationID); err != nil {
		log.Println(err)
	}

	if teamOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Команда не найдена"))
		return
	}

	if result := tx.Exec("DELETE FROM teams_on_moderation WHERE id = ?", teamOnModerationID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при удалении команды на модерации"))
		return
	}

	var filter bson.M
	if teamID.Valid {
		filter = bson.M{"team_id": teamID}
	} else {
		filter = bson.M{"team_on_moderation_id": teamOnModerationID}
	}

	if _, err := h.TeamsOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при удалении обложки команды"))
		return
	}

	tx.Commit()

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обращение на модерацию команды перевода успешно отклонено"))
}
