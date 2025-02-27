package handlers

import (
	"log"
	"strings"
	"sync"

	"github.com/Araks1255/mangacage_service_bot/pkg/common/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) Login(update tgbotapi.Update, mutex *sync.RWMutex) {
	tgUserID := update.Message.Chat.ID

	rawArgs := update.Message.CommandArguments()
	if rawArgs == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите имя пользователя и пароль через пробел после вызова команды.\n\nПример: /login user_name password"))
		return
	}
	args := strings.Fields(rawArgs)

	userName := args[0]
	password := args[1]

	var userRoles []string
	h.DB.Raw("SELECT roles.name FROM roles INNER JOIN user_roles ON roles.id = user_roles.role_id INNER JOIN users ON user_roles.user_id = users.id WHERE users.user_name = ?", userName).Scan(&userRoles)

	if !utils.DoesUserHaveRequiredRole(userRoles) {
		msg := tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором (возможно, в имени или пароле опечатка)")
		h.Bot.Send(msg)
		return
	}

	var (
		hashPassword string
		userID       uint
	)
	row := h.DB.Raw("SELECT password, id FROM users WHERE user_name = ?", userName).Row()
	if err := row.Scan(&hashPassword, &userID); err != nil {
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, err.Error()))
		return
	}

	if !utils.IsPasswordCorrect(password, hashPassword) {
		msg := tgbotapi.NewMessage(tgUserID, "Неверный пароль")
		h.Bot.Send(msg)
		return
	}

	if result := h.DB.Exec("UPDATE users SET tg_user_id = ? WHERE user_name = ?", tgUserID, userName); result.Error != nil {
		log.Println(result.Error)
		msg := tgbotapi.NewMessage(tgUserID, result.Error.Error())
		h.Bot.Send(msg)
		return
	}

	mutex.Lock()
	h.AllowedUsers[tgUserID] = userID
	mutex.Unlock()

	msg := tgbotapi.NewMessage(tgUserID, "Вход в аккаунт выполнен успешно")
	h.Bot.Send(msg)
}
