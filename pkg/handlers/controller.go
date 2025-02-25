package handlers

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type handler struct {
	Bot               *tgbotapi.BotAPI
	DB                *gorm.DB
	AllowedUsersTgIds map[int64]uint
}

func RegisterCommands(bot *tgbotapi.BotAPI, db *gorm.DB) {
	h := handler{
		Bot: bot,
		DB:  db,
	}

	h.AllowedUsersTgIds = make(map[int64]uint, 10) // Потом надо будет заполнять из бд автоматически при запуске

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	var mutex sync.RWMutex

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Приветствую!\n\nЭтот бот предназначен для модерации манги на сайте mangacage.\n\nЧтобы войти в аккаунт, вызовите функцию /login, указав через пробел своё имя пользователя и пароль.\n\n(Пример: /login user_name password)")
				bot.Send(msg)

			case "login":
				go h.Login(update, &mutex)

			case "get_titles_on_moderation":
				go h.GetTitlesOnModeration(update)

			case "approve_title":
				go h.ApproveTitle(update)

			case "return_title_to_moderation":
				go h.ReturnTitleToModeration(update)

			case "get_chapters_on_moderation":
				go h.GetChaptersOnModeration(update)

			}
		}
	}
}
