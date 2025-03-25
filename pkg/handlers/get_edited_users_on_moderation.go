package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetEditedUsersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не модератор"))
		return
	}

	var editedUsers []struct {
		gorm.Model
		UserName         string
		AboutYourself    string
		ExistingUserName string
	}

	h.DB.Raw(
		`SELECT u.id, u.created_at, u.user_name, u.about_yourself,
		users.user_name AS existing_user_name
		FROM users_on_moderation AS u
		INNER JOIN users ON users.id = u.existing_id`,
	).Scan(&editedUsers)

	if len(editedUsers) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет пользователей, ожидающих подтверждения изменений аккаунта"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(editedUsers); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\nИмя пользователя (на данный момент): %s\n\nИзменения:\n Имя: %s\n О себе: %s\n\nИзменения внесены:\n%s",
			editedUsers[i].ID, editedUsers[i].ExistingUserName, editedUsers[i].UserName, editedUsers[i].AboutYourself, editedUsers[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}
	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы подробнее увидеть изменения (со старыми данными и аватаркой), вызовите команду review_user с указанием айди нужного обращения\n\nПример: /review_user 2"))
}
