package handlers

import (
	"log"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ReturnTitleToModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsersTgIds[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	var userRoles []string
	h.DB.Raw("SELECT roles.name FROM roles INNER JOIN user_roles ON roles.id = user_roles.role_id INNER JOIN users ON user_roles.user_id = users.id WHERE users.id = ?", userID).Scan(&userRoles)

	if IsUserAdmin := slices.Contains(userRoles, "admin"); !IsUserAdmin {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Возвращать тайтлы на модерацию могут только администраторы"))
		return
	}

	desiredTitleName := update.Message.CommandArguments()
	if desiredTitleName == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название тайтла, который хотите вернуть на модерацию, через пробел после команды\n\nПример: /return_title_to_moderation Мёртвый аккаунт"))
		return
	}

	if result := h.DB.Exec("UPDATE titles SET on_moderation = true, moderator_id = ? WHERE name = ?", userID, desiredTitleName); result.RowsAffected == 0 {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось вернуть тайтл на модерацию. Возможно, была допущена опечатка"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно возвращён на модерацию"))
}
