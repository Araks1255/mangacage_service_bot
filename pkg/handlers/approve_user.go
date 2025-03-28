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

func (h handler) ApproveUser(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или админом"))
		return
	}

	desiredUserOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения пользователя, обращение которого хотите одобрить\n\nПример: /approve_user 2"))
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

	var userOnModeration models.UserOnModeration
	tx.Raw("SELECT * FROM users_on_moderation WHERE id = ?", desiredUserOnModerationID).Scan(&userOnModeration)
	if userOnModeration.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "На модерации нет пользователя с таким id обращения"))
		return
	}

	var user models.User
	tx.Raw("SELECT * FROM users WHERE id = ?", userOnModeration.ExistingID).Scan(&user)

	doesUserExist := user.ID != 0

	editUser(userOnModeration, &user)

	if result := tx.Save(&user); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании или обновлении пользователя"))
		return
	}

	var filter bson.M
	if doesUserExist {
		filter = bson.M{"user_id": user.ID}

		var userProfilePicture struct {
			UserID         uint   `bson:"user_id"`
			ProfilePicture []byte `bson:"profile_picture"`
		}

		if err := h.UsersOnModerationProfilePictures.FindOne(context.TODO(), filter).Decode(&userProfilePicture); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске аватарки пользователя"))
			return
		}

		opts := options.Update().SetUpsert(true)
		if _, err = h.UsersProfilePictures.UpdateOne(context.TODO(), userProfilePicture, opts); err != nil {
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при обновлении аватарки пользователя"))
			return
		}
	}

	tx.Commit()

	if doesUserExist {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Аккаунт пользователя успешно обновлен"))
	} else {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Аккаунт пользователя успешно верифицирован"))
	}

	if _, err = h.UsersOnModerationProfilePictures.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
	}

	if result := h.DB.Exec("DELETE FROM users_on_moderation WHERE id = ?", userOnModeration.ID); result.Error != nil {
		log.Println(result.Error)
	}

	// Добавить уведомление о верификации аккаунта (возможно, когда-то потом)
}

func editUser(userOnModeration models.UserOnModeration, user *models.User) {
	if userOnModeration.UserName.String != "" {
		user.UserName = userOnModeration.UserName.String
	}
	if userOnModeration.AboutYourself != "" {
		user.AboutYourself = userOnModeration.AboutYourself
	}
	if userOnModeration.Password != "" {
		user.Password = userOnModeration.Password
	}
}
