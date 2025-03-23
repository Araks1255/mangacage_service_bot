package handlers

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type handler struct {
	Bot                       *tgbotapi.BotAPI
	DB                        *gorm.DB
	TitlesOnModerationCovers  *mongo.Collection
	TitlesCovers              *mongo.Collection
	ChaptersOnModerationPages *mongo.Collection
	ChaptersPages             *mongo.Collection
	AllowedUsers              map[int64]uint
}

func RegisterCommands(bot *tgbotapi.BotAPI, client *mongo.Client, db *gorm.DB) {
	titlesOnModerationCovers := client.Database("mangacage").Collection("titles_on_moderation_covers")
	titlesCovers := client.Database("mangacage").Collection("titles_covers")

	chaptersOnModerationPages := client.Database("mangacage").Collection("chapters_on_moderation_pages")
	chaptersPagesCollection := client.Database("mangacage").Collection("chapters_pages")

	h := handler{
		Bot:                       bot,
		DB:                        db,
		TitlesOnModerationCovers:  titlesOnModerationCovers,
		TitlesCovers:              titlesCovers,
		ChaptersOnModerationPages: chaptersOnModerationPages,
		ChaptersPages:             chaptersPagesCollection,
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

			case "get_new_titles_on_moderation":
				go h.GetNewTitlesOnModeration(update)

			case "get_edited_titles_on_moderation":
				go h.GetEditedTitlesOnModeration(update)

			case "get_new_volumes_on_moderation":
				go h.GetNewVolumesOnModeration(update)

			case "get_edited_volumes_on_moderation":
				go h.GetEditedVolumesOnModeration(update)

			case "get_new_chapters_on_moderation":
				go h.GetNewChaptersOnModeration(update)

			case "get_edited_chapters_on_moderation":
				go h.GetEditedChaptersOnModeration(update)

			case "approve_title":
				go h.ApproveTitle(update)

			case "review_chapter":
				go h.ReviewChapter(update)

			case "review_title":
				go h.ReviewTitle(update)

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
