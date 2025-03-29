package handlers

import (
	"context"
	"log"
	"strconv"

	"github.com/Araks1255/mangacage/pkg/common/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h handler) ApproveTeam(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredTeamOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Укажите id обращения желаемой команды перевода после вызова команды\n\nПример: /approve_team 2"))
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

	var teamOnModeration models.TeamOnModeration
	tx.Raw("SELECT * FROM teams_on_moderation WHERE id = ?", desiredTeamOnModerationID).Scan(&teamOnModeration)
	if teamOnModeration.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Команда не надена"))
		return
	}

	doesTeamExist := teamOnModeration.ExistingID.Int64 != 0

	var team models.Team

	if doesTeamExist {
		tx.Raw("SELECT * FROM teams WHERE id = ?", teamOnModeration.ExistingID.Int64).Scan(&team)
	}

	log.Println(team)

	EditTeam(teamOnModeration, &team)
	team.ModeratorID = userID

	if result := tx.Save(&team); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании или редактировании команды"))
		return
	}

	session, err := h.MongoClient.StartSession()
	if err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка при создании сессии"))
		return
	}
	defer session.EndSession(context.TODO())

	if !doesTeamExist {
		var teamOnModerationCover struct {
			TeamOnModerationID uint   `bson:"team_on_moderation"`
			Cover              []byte `bson:"cover"`
		}

		filter := bson.M{"team_on_moderation_id": teamOnModeration.ID}

		if err := h.TeamsOnModerationCovers.FindOne(context.TODO(), filter).Decode(&teamOnModerationCover); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске обложки команды"))
			return
		}

		var teamCover struct {
			TeamID uint   `bson:"team_id"`
			Cover  []byte `bson:"cover"`
		}

		teamCover.TeamID = team.ID
		teamCover.Cover = teamOnModerationCover.Cover

		if _, err := h.TeamsCovers.InsertOne(context.TODO(), teamCover); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании обложки команды"))
			return
		}

		if result := tx.Exec("UPDATE users SET team_id = ? WHERE id = ?", team.ID, teamOnModeration.CreatorID); result.Error != nil {
			log.Println(result.Error)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось присоеденить создателя к его команде"))
			return
		}

		if result := tx.Exec(
			"INSERT INTO user_roles (user_id, role_id) VALUES (?, (SELECT id FROM roles WHERE name = 'team_leader'))",
			teamOnModeration.CreatorID,
		); result.Error != nil {
			log.Println(result.Error)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось назначить создателя команды её лидером"))
			return
		}

		tx.Commit()

		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Команда успешно создана и её создатель назначен её лидером"))

		if _, err := h.TeamsOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить ненужную обложку"))
		}

		// Уведомление создателю

		return
	}

	var teamOnModerationCover struct {
		TeamOnModerationID uint   `bson:"team_on_moderation_id"`
		Cover              []byte `bson:"cover"`
	}

	filter := bson.M{"team_id": team.ID}

	if err := h.TeamsOnModerationCovers.FindOne(context.TODO(), filter).Decode(&teamOnModerationCover); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось найти обложку команды на модерации"))
		return
	}

	coverUpdate := bson.M{"$set": bson.M{"cover": teamOnModerationCover.Cover}}
	opts := options.Update().SetUpsert(true)

	if _, err := h.TeamsCovers.UpdateOne(context.TODO(), filter, coverUpdate, opts); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось обновить обложку команды"))
		return
	}

	tx.Commit()

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Команда перевода успешно обновлена"))

	if _, err := h.TeamsOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось удалить ненужную обложку"))
	}
	// Уведомление создателю
}

func EditTeam(teamOnModeration models.TeamOnModeration, team *models.Team) {
	if teamOnModeration.Name != "" {
		team.Name = teamOnModeration.Name
	}
	if teamOnModeration.Description != "" {
		team.Description = teamOnModeration.Description
	}

	if !teamOnModeration.ExistingID.Valid {
		team.CreatorID = teamOnModeration.CreatorID
	}
}
