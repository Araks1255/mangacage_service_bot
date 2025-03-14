package handlers

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type handler struct {
	Bot          *tgbotapi.BotAPI
	DB           *gorm.DB
	Collection   *mongo.Collection
	AllowedUsers map[int64]uint
}

func RegisterCommands(bot *tgbotapi.BotAPI, client *mongo.Client, db *gorm.DB) {
	chapterPagesCollection := client.Database("mangacage").Collection("chapters_pages")

	h := handler{
		Bot:        bot,
		DB:         db,
		Collection: chapterPagesCollection,
	}

	h.AllowedUsers = make(map[int64]uint, 10)

	var allowedUsersIds []uint
	h.DB.Raw("SELECT users.id FROM users INNER JOIN user_roles ON users.id = user_roles.user_id INNER JOIN roles ON user_roles.role_id = roles.id WHERE roles.name = 'moder' OR roles.name = 'admin'").Scan(&allowedUsersIds)

	var allowedUsersTgIDs []int64
	h.DB.Raw("SELECT users.tg_user_id FROM users INNER JOIN user_roles ON users.id = user_roles.user_id INNER JOIN roles ON user_roles.role_id = roles.id WHERE roles.name = 'moder' OR roles.name = 'admin'").Scan(&allowedUsersTgIDs)

	for i := 0; i < len(allowedUsersIds); i++ {
		h.AllowedUsers[allowedUsersTgIDs[i]] = allowedUsersIds[i]
	}

	log.Println(h.AllowedUsers)

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

			case "get_volumes_on_moderation":
				go h.GetVolumesOnModeration(update)

			case "approve_title":
				go h.ApproveTitle(update)

			case "return_title_to_moderation":
				go h.ReturnTitleToModeration(update)

			case "get_chapters_on_moderation":
				go h.GetChaptersOnModeration(update)

			case "review_chapter":
				go h.ReviewChapter(update)

			case "approve_chapter":
				go h.ApproveChapter(update)

			case "reject_title":
				go h.RejectTitle(update)

			case "reject_chapter":
				go h.RejectChapter(update)

			case "approve_volume":
				go h.ApproveVolume(update)
			}
		}
	}
}
