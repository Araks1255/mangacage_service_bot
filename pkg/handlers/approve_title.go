package handlers

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/Araks1255/mangacage/pkg/common/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

type TitleCover struct {
	TitleID uint   `bson:"title_id"`
	Cover   []byte `bson:"cover"`
}

func (h handler) ApproveTitle(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	titleOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения тайтла, который хотите одобрить\n\nПример: /approve_title 2"))
		return
	}

	var titleOnModeration models.TitleOnModeration
	h.DB.Raw("SELECT * FROM titles_on_moderation WHERE id = ?", titleOnModerationID).Scan(&titleOnModeration)
	if titleOnModeration.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "На модерации нет тайтла с таким id обращения"))
		return
	}

	var title models.Title
	h.DB.Raw("SELECT * FROM titles WHERE id = ?", titleOnModeration.ExistingID).Scan(&title)

	doesTitleExist := title.ID != 0

	ConvertToTitle(titleOnModeration, &title)
	title.ModeratorID = sql.NullInt64{Int64: int64(userID), Valid: true}

	tx := h.DB.Begin()

	if title.ID != 0 {
		if result := tx.Exec("DELETE FROM title_genres WHERE title_id = ?", title.ID); result.Error != nil {
			tx.Rollback()
			log.Println(result.Error)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при удалении старых жанров тайтла"))
			return
		}
	}

	if result := tx.Save(&title); result.Error != nil {
		tx.Rollback()
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании или обновлении тайтла"))
		return
	}

	if err = AddGenresToTitle(title.ID, titleOnModeration.Genres, tx); err != nil {
		tx.Rollback()
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при добавлении жанров тайтлу"))
		return
	}

	var titleCover TitleCover

	filter := bson.M{"title_id": titleOnModerationID}
	if err = h.TitlesOnModerationCovers.FindOne(context.TODO(), filter).Decode(&titleCover); err != nil {
		tx.Rollback()
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске обложки тайтла"))
		return
	}
	
	titleCover.TitleID = title.ID
	
	if _, err = h.TitlesCovers.InsertOne(context.TODO(), titleCover); err != nil {
		tx.Rollback()
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при вставке обложки тайтла"))
		return
	}
	
	tx.Commit()
	
	if doesTitleExist {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно изменён"))
	} else {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Тайтл успешно создан"))
	}
		
	if _, err = h.TitlesOnModerationCovers.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
	}

	if result := h.DB.Exec("DELETE FROM titles_on_moderation WHERE id = ?", titleOnModeration.ID); result.Error != nil {
		log.Println(result.Error)
	}
}

func ConvertToTitle(titleOnModeration models.TitleOnModeration, title *models.Title) {
	title.Name = titleOnModeration.Name
	title.Description = titleOnModeration.Description
	title.AuthorID = titleOnModeration.AuthorID
	title.CreatorID = titleOnModeration.CreatorID
	title.TeamID = titleOnModeration.TeamID
}

func AddGenresToTitle(titleID uint, genres []string, tx *gorm.DB) error {
	query := `
		INSERT INTO title_genres (title_id, genre_id)
		SELECT ?, genres.id
		FROM genres
		JOIN UNNEST(?::TEXT[]) AS genre_name ON genres.name = genre_name
	`

	if result := tx.Exec(query, titleID, pq.Array(genres)); result.Error != nil {
		log.Println(result.Error)
		return result.Error
	}

	return nil
}
