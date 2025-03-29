package handlers

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type handler struct {
	Bot *tgbotapi.BotAPI
	DB  *gorm.DB

	MongoClient *mongo.Client

	TitlesOnModerationCovers *mongo.Collection
	TitlesCovers             *mongo.Collection

	ChaptersOnModerationPages *mongo.Collection
	ChaptersPages             *mongo.Collection

	UsersOnModerationProfilePictures *mongo.Collection
	UsersProfilePictures             *mongo.Collection

	TeamsOnModerationCovers *mongo.Collection
	TeamsCovers             *mongo.Collection

	AllowedUsers map[int64]uint
}

func RegisterCommands(bot *tgbotapi.BotAPI, client *mongo.Client, db *gorm.DB) {
	mangacageDB := client.Database("mangacage")

	titlesOnModerationCovers := mangacageDB.Collection("titles_on_moderation_covers")
	titlesCovers := mangacageDB.Collection("titles_covers")

	chaptersOnModerationPages := mangacageDB.Collection("chapters_on_moderation_pages")
	chaptersPages := mangacageDB.Collection("chapters_pages")

	usersOnModerationProfilePictures := mangacageDB.Collection("users_on_moderation_profile_pictures")
	usersProfilePictures := mangacageDB.Collection("users_profile_pictures")

	teamsOnModerationCovers := mangacageDB.Collection("teams_on_moderation_covers")
	teamsCovers := mangacageDB.Collection("teams_covers")

	h := handler{
		Bot:                              bot,
		DB:                               db,
		MongoClient:                      client,
		TitlesOnModerationCovers:         titlesOnModerationCovers,
		TitlesCovers:                     titlesCovers,
		ChaptersOnModerationPages:        chaptersOnModerationPages,
		ChaptersPages:                    chaptersPages,
		UsersOnModerationProfilePictures: usersOnModerationProfilePictures,
		UsersProfilePictures:             usersProfilePictures,
		TeamsOnModerationCovers:          teamsOnModerationCovers,
		TeamsCovers:                      teamsCovers,
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

			case "get_new_users_on_moderation":
				go h.GetNewUsersOnModeration(update)

			case "get_new_teams_on_moderation":
				go h.GetNewTeamsOnModeration(update)

			case "get_edited_volumes_on_moderation":
				go h.GetEditedVolumesOnModeration(update)

			case "get_new_chapters_on_moderation":
				go h.GetNewChaptersOnModeration(update)

			case "get_edited_chapters_on_moderation":
				go h.GetEditedChaptersOnModeration(update)

			case "get_edited_users_on_moderation":
				go h.GetEditedUsersOnModeration(update)

			case "get_edited_teams_on_moderation":
				go h.GetEditedTeamsOnModeration(update)

			case "approve_title":
				go h.ApproveTitle(update)

			case "review_chapter":
				go h.ReviewChapter(update)

			case "review_title":
				go h.ReviewTitle(update)

			case "review_volume":
				go h.ReviewVolume(update)

			case "review_user":
				go h.ReviewUser(update)

			case "review_team":
				go h.ReviewTeam(update)

			case "approve_chapter":
				go h.ApproveChapter(update)

			case "approve_team":
				go h.ApproveTeam(update)

			case "approve_volume":
				go h.ApproveVolume(update)

			case "reject_title":
				go h.RejectTitle(update)

			case "reject_chapter":
				go h.RejectChapter(update)
			
			case "reject_team":
				go h.RejectTeam(update)
				
			}
		}
	}
}
