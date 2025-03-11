package handlers

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) RejectChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	_, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	desiredChapterID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id главы, которую хотите отвергнуть, после вызова функции\n\nПример: /reject_chapter 12"))
		return
	}

	var existingChapterName string
	h.DB.Raw("SELECT name FROM chapters WHERE id = ?", desiredChapterID).Scan(&existingChapterName)
	if existingChapterName == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	if result := h.DB.Exec("DELETE FROM chapters WHERE id = ?", desiredChapterID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава успешно отвергнута"))
}
