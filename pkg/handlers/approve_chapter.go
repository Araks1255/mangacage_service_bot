package handlers

import (
	"context"
	"log"
	"strings"

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

	desiredChapter := strings.ToLower(update.Message.CommandArguments())
	if desiredChapter == "" {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите название главы, которую хотите одобрить через пробел после вызова команды\n\nПример: /approve_chapter Глава 1"))
		return
	}

	if result := h.DB.Exec("UPDATE chapters SET on_moderation = false, moderator_id = ? WHERE name = ?", userID, desiredChapter); result.RowsAffected == 0 {
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Не удалось снять главу с модерации. Скорее всего вы совершили опечатку в названии"))
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

	if _, err := client.NotifyAboutReleaseOfNewChapterInTitle(context.Background(), &pb.ReleasedChapter{Name: desiredChapter}); err != nil {
		log.Println(err)
	}
}
