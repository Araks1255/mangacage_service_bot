package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func (h handler) ReviewTeam(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не модератор"))
		return
	}

	desiredTeamOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди обращения команды перевода, которую хотите рассмотреть, после вызова команды\n\nПример: /review_team 2"))
		return
	}

	var teamOnModerationID uint
	h.DB.Raw("SELECT id FROM teams_on_moderation WHERE id = ?", desiredTeamOnModerationID).Scan(&teamOnModerationID)
	if teamOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Команда перевода с таким id обращения не найдена"))
		return
	}

	var teamOnModeration struct {
		ID          uint
		CreatedAt   time.Time
		Name        string
		Description string
		ExistingID  uint
		Creator     string
	}

	h.DB.Raw(
		`SELECT t.id, t.created_at, t.name, t.description, t.existing_id,
		users.user_name AS creator
		FROM teams_on_moderation AS t
		INNER JOIN users ON users.id = t.creator_id
		WHERE t.id = ?`, teamOnModerationID,
	).Scan(&teamOnModeration)

	doesTeamExist := teamOnModeration.ExistingID != 0

	var response string

	if !doesTeamExist {
		var teamCover struct {
			TeamOnModerationID uint   `bson:"team_on_moderation_id"`
			Cover              []byte `bson:"cover"`
		}

		filter := bson.M{"team_on_moderation_id": teamOnModeration.ID}

		if err := h.TeamsOnModerationCovers.FindOne(context.TODO(), filter).Decode(&teamCover); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске обложки команды перевода"))
		}

		response = fmt.Sprintf(
			"Причина обращения: создание\nid обращения: %d\n\nНазвание: %s\nОписание: %s\nСоздатель: %s\n\nОтправлена на модерацию: %s",
			teamOnModeration.ID, teamOnModeration.Name, teamOnModeration.Description, teamOnModeration.Creator, teamOnModeration.CreatedAt.Format(time.DateTime),
		)

		cover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
			Name:  "cover",
			Bytes: teamCover.Cover,
		})
		cover.Caption = response

		h.Bot.Send(cover)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тут будут инструкции по одобрению и неодобрению"))
		return
	}

	var team struct {
		ID          uint
		CreatedAt   time.Time
		Name        string
		Description string
		Leader      string
		Moder       string
	}

	h.DB.Raw(
		`SELECT teams.id, teams.created_at, teams.name, teams.description,
		users.user_name AS leader, moders.user_name AS moder
		FROM teams
		LEFT JOIN users AS moders ON moders.id = teams.moderator_id
		INNER JOIN users ON teams.id = users.team_id
		INNER JOIN user_roles ON users.id = user_roles.user_id
		INNER JOIN roles ON user_roles.role_id = roles.id
		WHERE teams.id = ?
		AND roles.name = 'team_leader'`, teamOnModeration.ExistingID,
	).Scan(&team)

	response = fmt.Sprintf(
		"Причина обращения: редактирование\nid обращения: %d\n\nИнформация о команде (на данный момент):\n Название: %s\n Описание: %s\n\nЛидер: %s\nПоследний редактировавший модератор: %s\n\nСоздана:\n%s",
		teamOnModeration.ID, team.Name, team.Description, team.Leader, team.Moder, team.CreatedAt.Format(time.DateTime),
	)
	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	response = "Изменения:\n\n"

	if teamOnModeration.Name != "" {
		response += fmt.Sprintf("Название с \"%s\" на \"%s\"\n", team.Name, teamOnModeration.Name)
	}
	if teamOnModeration.Description != "" {
		response += fmt.Sprintf("Описание с \"%s\" на \"%s\"", team.Description, teamOnModeration.Description)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	var newTeamCover struct {
		TeamID uint   `bson:"team_id"`
		Cover  []byte `bson:"cover"`
	}

	filter := bson.M{"team_id": team.ID}

	if err := h.TeamsOnModerationCovers.FindOne(context.TODO(), filter).Decode(&newTeamCover); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Обложка не обновлена (или произошла ошибка)"))
		return
	}

	newCover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
		Name:  "new_cover",
		Bytes: newTeamCover.Cover,
	})
	newCover.Caption = "Новая обложка"
	h.Bot.Send(newCover)

	var oldTeamCover struct { // Тут можно было бы использовать одну переменную для двух обложек, но это бы было довольно нечитаемо
		TeamID uint   `bson:"team_id"`
		Cover  []byte `bson:"cover"`
	}

	if err := h.TeamsCovers.FindOne(context.TODO(), filter).Decode(&oldTeamCover); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось получить нынешнюю обложку команды"))
		return
	}

	oldCover := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
		Name:  "old_cover",
		Bytes: oldTeamCover.Cover,
	})
	oldCover.Caption = "Старая (нынешняя) обложка"
	h.Bot.Send(oldCover)
	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Инструкции"))
}
