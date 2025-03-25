package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

func (h handler) ReviewUser(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не модератор"))
		return
	}

	desiredUserOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите айди обращения пользователя, аккаунт которого хотите рассмотреть, после вызова команды\n\nПример: /review_user 2"))
		return
	}

	var existingUserOnModerationID uint
	h.DB.Raw("SELECT id FROM users_on_moderation WHERE id = ?", desiredUserOnModerationID).Scan(&existingUserOnModerationID)
	if existingUserOnModerationID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Пользователь с таким id обращения не найден"))
		return
	}

	var userOnModeration struct {
		gorm.Model
		UserName      string
		AboutYourself string
		ExistingID    uint
	}

	h.DB.Raw("SELECT id, created_at, user_name, about_yourself, existing_id FROM users_on_moderation WHERE id = ?", existingUserOnModerationID).Scan(&userOnModeration)

	isUserNew := userOnModeration.ExistingID == 0

	var response string

	if isUserNew {
		response = fmt.Sprintf(
			"Причина обращения: верификация аккаунта\nid обращения: %d\n\nИмя: %s\nО себе: %s\n\nЗарегистрирован:\n%s",
			userOnModeration.ID, userOnModeration.UserName, userOnModeration.AboutYourself, userOnModeration.CreatedAt.Format(time.DateTime),
		)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы верифицировать аккаунт пользователя, вызовите команду approve_user с указанием id его обращения\nЧтобы отклонить обращение на верификацию аккаунта, вызовите команду reject_user с указанием id обращения\n\nПримеры:\n/approve_user 2\n/reject_user 2"))
		return
	}

	var user struct {
		gorm.Model
		UserName     string
		AboutYorself string
	}

	h.DB.Raw("SELECT id, created_at, user_name, about_yourself FROM users WHERE id = ?", userOnModeration.ExistingID).Scan(&user)

	response = fmt.Sprintf(
		"Причина обращения: редактирование аккаунта\nid обращения: %d\n\nИнформация об аккаунте (на данный момент)\n Имя: %s\n О себе: %s\n\nЗарегистрирован:\n%s",
		userOnModeration.ID, user.UserName, user.AboutYorself, user.CreatedAt.Format(time.DateTime),
	)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	response = "Изменения:\n\n"

	if userOnModeration.UserName != "" {
		response += fmt.Sprintf("Имя пользователя с \"%s\" на \"%s\"\n\n", user.UserName, userOnModeration.UserName)
	}
	if userOnModeration.AboutYourself != "" {
		response += fmt.Sprintf("Раздел \"о себе\" c\n\"%s\"\nна\n\"%s\"", user.AboutYorself, userOnModeration.AboutYourself)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, response))

	var newUserProfilePicture struct {
		UserID         uint   `bson:"user_id"`
		ProfilePicture []byte `bson:"profile_picture"`
	}

	filter := bson.M{"user_id": user.ID}

	if err := h.UsersOnModerationProfilePictures.FindOne(context.TODO(), filter).Decode(&newUserProfilePicture); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Аватарка не обновлена"))
		return
	}

	var oldUserProfilePicture struct {
		UserID         uint   `bson:"user_id"`
		ProfilePicture []byte `bson:"profile_picture"`
	}

	if err := h.UsersProfilePictures.FindOne(context.TODO(), filter).Decode(&oldUserProfilePicture); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Старая аватарка не найдена (скорее всего, её и не было)"))
	} else {
		oldProfilePicture := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
			Name:  "profilePicture",
			Bytes: oldUserProfilePicture.ProfilePicture,
		})
		oldProfilePicture.Caption = "Старая аватарка"
		h.Bot.Send(oldProfilePicture)
	}

	newProfilePicture := tgbotapi.NewPhoto(tgUserID, tgbotapi.FileBytes{
		Name:  "profilePicture",
		Bytes: newUserProfilePicture.ProfilePicture,
	})
	newProfilePicture.Caption = "Новая аватарка"

	h.Bot.Send(newProfilePicture)

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "тут будут инструкции по одобрению и не одобрению"))
}
