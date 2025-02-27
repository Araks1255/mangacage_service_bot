package main

import (
	"github.com/Araks1255/mangacage_service_bot/pkg/common/db"
	"github.com/Araks1255/mangacage_service_bot/pkg/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

var allowedUsersTgIds map[int64]uint

func main() {
	viper.SetConfigFile("./pkg/common/envs/.env")
	viper.ReadInConfig()

	token := viper.Get("TOKEN").(string)
	dbUrl := viper.Get("DB_URL").(string)

	db, err := db.Init(dbUrl)
	if err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	handlers.RegisterCommands(bot, db)
}
