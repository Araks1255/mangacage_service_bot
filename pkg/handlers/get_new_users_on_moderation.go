package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetNewUsersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не модератор"))
		return
	}

	var newUsers []struct {
		gorm.Model
		UserName      string
		AboutYourself string
	}

	h.DB.Raw(
		`SELECT id, created_at, user_name, about_yourself
		FROM users_on_moderation WHERE existing_id IS NULL`,
	).Scan(&newUsers)

	if len(newUsers) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Нет пользователей ожидающих верификации"))
		return
	}

	var (
		response string
		msg      tgbotapi.MessageConfig
	)

	for i := 0; i < len(newUsers); i++ {
		response = fmt.Sprintf(
			"id обращения: %d\n\nИмя пользователя: %s\nО себе: %s\n\nЗарегистрирован:\n%s",
			newUsers[i].ID, newUsers[i].UserName, newUsers[i].AboutYourself, newUsers[i].CreatedAt.Format(time.DateTime),
		)
		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Чтобы отдельно рассмотреть аккаунт пользователя, ожидающего верификации, укажите id его обращения после вызова команды review_user\n\nПример: /review_user 2"))
}
