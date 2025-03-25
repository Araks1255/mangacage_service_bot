package handlers

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/Araks1255/mangacage/pkg/common/models"
	pb "github.com/Araks1255/mangacage_protos"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChapterPages struct {
	ChapterID uint
	Pages     [][]byte
}

func (h handler) ApproveChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или админом"))
		return
	}

	chapterOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения главы, которую хотите одобрить\n\nПример: /approve_chapter 2"))
		return
	}

	var chapterOnModeration models.ChapterOnModeration
	h.DB.Raw("SELECT * FROM chapters_on_moderation WHERE id = ?", chapterOnModerationID).Scan(&chapterOnModeration)
	if chapterOnModeration.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "На модерации нет главы с таким id обращения"))
		return
	}

	var chapter models.Chapter
	h.DB.Raw("SELECT * FROM chapters WHERE id = ?", chapterOnModeration.ExistingID).Scan(&chapter)

	doesChapterExist := chapter.ID != 0

	EditChapter(chapterOnModeration, &chapter)
	chapter.ModeratorID = sql.NullInt64{Int64: int64(userID), Valid: true}

	tx := h.DB.Begin()

	if result := tx.Save(&chapter); result.Error != nil {
		tx.Rollback()
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании или обновлении главы"))
		return
	}

	var filter bson.M

	if !doesChapterExist {
		var chapterPages ChapterPages

		filter = bson.M{"chapter_id": chapterOnModeration.ID}

		if err := h.ChaptersOnModerationPages.FindOne(context.TODO(), filter).Decode(&chapterPages); err != nil {
			tx.Rollback()
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при поиске страниц главы"))
			return
		}

		chapterPages.ChapterID = chapter.ID

		if _, err := h.ChaptersPages.InsertOne(context.TODO(), chapterPages); err != nil {
			tx.Rollback()
			log.Println(err)
			h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при вставке страниц главы"))
			return
		}
	}

	tx.Commit()

	if doesChapterExist {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава успешно изменена"))
	} else {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава успешно создана"))
	}

	if _, err := h.ChaptersOnModerationPages.DeleteOne(context.TODO(), filter); err != nil {
		log.Println(err)
	}

	if result := h.DB.Exec("DELETE FROM chapters_on_moderation WHERE id = ?", chapterOnModeration.ID); result.Error != nil {
		log.Println(result.Error)
	}

	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewNotificationsClient(conn)

	if !doesChapterExist {
		if _, err := client.NotifyAboutReleaseOfNewChapterInTitle(context.TODO(), &pb.ReleasedChapter{ID: uint32(chapter.ID)}); err != nil {
			log.Println(err)
		}
	}
}

func EditChapter(chapterOnModeration models.ChapterOnModeration, chapter *models.Chapter) {
	if chapterOnModeration.Name != "" {
		chapter.Name = chapterOnModeration.Name
	}
	if chapterOnModeration.Description != "" {
		chapter.Description = chapterOnModeration.Description
	}
	if chapterOnModeration.NumberOfPages != 0 {
		chapter.NumberOfPages = chapterOnModeration.NumberOfPages
	}
	if chapterOnModeration.VolumeID != 0 {
		chapter.VolumeID = chapterOnModeration.VolumeID
	}
	if chapterOnModeration.CreatorID != 0 {
		chapter.CreatorID = chapterOnModeration.CreatorID
	}
	if chapterOnModeration.VolumeID != 0 {
		chapter.VolumeID = chapterOnModeration.VolumeID
	}
}
