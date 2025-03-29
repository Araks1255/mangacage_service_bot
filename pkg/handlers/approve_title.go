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
	"go.mongodb.org/mongo-driver/mongo/options"
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

	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var titleOnModeration models.TitleOnModeration
	tx.Raw("SELECT * FROM titles_on_moderation WHERE id = ?", titleOnModerationID).Scan(&titleOnModeration)
	if titleOnModeration.ID == 0 {
		tx.Rollback()
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "На модерации нет тайтла с таким id обращения"))
		return
	}

	var title models.Title
	tx.Raw("SELECT * FROM titles WHERE id = ?", titleOnModeration.ExistingID.Int64).Scan(&title)

	doesTitleExist := title.ID != 0

	ConvertToTitle(titleOnModeration, &title)
	title.ModeratorID = sql.NullInt64{Int64: int64(userID), Valid: true}

	if doesTitleExist {
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

	var filter bson.M
	if doesTitleExist {
		filter = bson.M{"title_id": title.ID}
	} else {
		filter = bson.M{"title_on_moderation_id": titleOnModeration.ID}
	}

	if err = h.TitlesOnModerationCovers.FindOne(context.TODO(), filter).Decode(&titleCover); err != nil {
		tx.Rollback()
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске обложки тайтла"))
		return
	}

	filter = bson.M{"title_id": title.ID}
	coverUpdate := bson.M{"$set": bson.M{"cover": titleCover.Cover}}
	opts := options.Update().SetUpsert(true)

	if _, err := h.TitlesCovers.UpdateOne(context.TODO(), filter, coverUpdate, opts); err != nil {
		tx.Rollback()
		log.Println(err)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании обложки тайтла"))
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
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при удалении ненужной обложки тайтла"))
	}

	if result := h.DB.Exec("DELETE FROM titles_on_moderation WHERE id = ?", titleOnModeration.ID); result.Error != nil {
		log.Println(result.Error)
	}
}

func ConvertToTitle(titleOnModeration models.TitleOnModeration, title *models.Title) {
	if titleOnModeration.Name != "" { // При редактировании неизменённые столбцы будут NULL, а сюда будет возвращаться zero value. На него и проверяю, чтобы изменить только то, что надо
		title.Name = titleOnModeration.Name
	}
	if titleOnModeration.Description != "" {
		title.Description = titleOnModeration.Description
	}
	if titleOnModeration.AuthorID.Int64 != 0 {
		title.AuthorID = uint(titleOnModeration.AuthorID.Int64)
	}

	if !titleOnModeration.ExistingID.Valid { // Если тайтл новый, то ставим ему id создателя такой же как у тайтла на модерации
		title.CreatorID = titleOnModeration.CreatorID
	}
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
