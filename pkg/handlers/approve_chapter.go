package handlers

import (
	"context"
	"log"
	"strconv"

	pb "github.com/Araks1255/mangacage_protos"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (h handler) ApproveChapter(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или админом"))
		return
	}

	desiredChapterID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id главы, которую хотите одобрить, после вызова функции\n\nПример: /approve_chapter 12"))
		return
	}

	var existingChapterName string
	h.DB.Raw("SELECT name FROM chapters WHERE id = ? AND on_moderation", desiredChapterID).Scan(&existingChapterName)
	if existingChapterName == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава не найдена"))
		return
	}

	if result := h.DB.Exec("UPDATE chapters SET on_moderation = false, moderator_id = ? WHERE id = ?", userID, desiredChapterID); result.Error != nil {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Ошибка сервера"))
		return
	}

	h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Глава успешно снята с модерации"))

	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewNotificationsClient(conn)

	if _, err := client.NotifyAboutReleaseOfNewChapterInTitle(context.Background(), &pb.ReleasedChapter{Name: existingChapterName}); err != nil {
		log.Println(err)
	}
}
