package grpc

import (
	"context"
	"fmt"

	pb "github.com/Araks1255/mangacage_service_protos"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedServiceNotificationsServer
}

var bot *tgbotapi.BotAPI
var db *gorm.DB

func InitBotAndDB(existingBot *tgbotapi.BotAPI, existingDB *gorm.DB) {
	bot = existingBot
	db = existingDB
}

func (s Server) NotifyAboutNewTitleOnModeration(ctx context.Context, title *pb.TitleOnModeration) (*pb.Empty, error) {
	var allowedUsersTgIds []int64
	db.Raw("SELECT users.tg_user_id FROM users INNER JOIN user_roles ON users.id = user_roles.user_id INNER JOIN roles ON user_roles.role_id = roles.id WHERE roles.name = 'moder' OR 'admin'").Scan(&allowedUsersTgIds)

	var msg tgbotapi.MessageConfig
	for i := 0; i < len(allowedUsersTgIds); i++ {
		msg = tgbotapi.NewMessage(allowedUsersTgIds[i], fmt.Sprintf("На модерацию пришёл новый тайтл\n\nНазвание: %s", title.TitleName))
		_, err := bot.Send(msg)
		if err != nil {
			return &pb.Empty{}, err
		}
	}

	return &pb.Empty{}, nil
}

func (s Server) NotifyAboutNewChapterOnModeration(ctx context.Context, chapter *pb.ChapterOnModeration) (*pb.Empty, error) {
	var allowedUsersTgIds []int64
	db.Raw("SELECT users.tg_user_id FROM users INNER JOIN user_roles ON users.id = user_roles.user_id INNER JOIN roles ON user_roles.role_id = roles.id WHERE roles.name = 'moder' OR 'admin'").Scan(&allowedUsersTgIds)

	var msg tgbotapi.MessageConfig
	for i := 0; i < len(allowedUsersTgIds); i++ {
		msg = tgbotapi.NewMessage(allowedUsersTgIds[i], fmt.Sprintf("На модерацию пришла новая глава для тайтла:\n%s\n\nНазвание:\n%s", chapter.TitleName, chapter.ChapterName))
		_, err := bot.Send(msg)
		if err != nil {
			return &pb.Empty{}, err
		}
	}

	return &pb.Empty{}, nil
}
