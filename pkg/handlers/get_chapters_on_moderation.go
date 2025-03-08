package handlers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func (h handler) GetChaptersOnModeration(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	if _, ok := h.AllowedUsers[tgUserID]; !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, `Вы не являетесь модератором или администратором`))
		return
	}

	type result struct {
		gorm.Model
		Name          string
		Description   string
		NumberOfPages int
		Volume        string
		Title         string
		Moder         string
	}

	var chapters []result
	h.DB.Raw(`SELECT chapters.id, chapters.created_at, chapters.updated_at, chapters.deleted_at, chapters.name, chapters.description,
		chapters.number_of_pages, volumes.name AS volume, titles.name AS title, users.id AS moder FROM chapters
		INNER JOIN volumes ON chapters.volume_id = volumes.id
		INNER JOIN titles ON volumes.title_id = titles.id
		LEFT JOIN users ON chapters.moderator_id = users.id
		WHERE chapters.on_moderation`).Scan(&chapters)

	if len(chapters) == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, `Нет глав на модерации`))
		return
	}

	var (
		msg      tgbotapi.MessageConfig
		response string
	)

	for i := 0; i < len(chapters); i++ {
		response = fmt.Sprintf(
			`ID главы: %d
			
			Глава тома %s тайтла %s
			
			Название: %s
			Описание: %s

			Количество страниц: %d

			Последний раз редактировавший модератор: %s
			
			Создана: %s
			Последний раз изменена: %s`,
			chapters[i].ID,
			chapters[i].Volume, chapters[i].Title,
			chapters[i].Name,
			chapters[i].Description,
			chapters[i].NumberOfPages,
			chapters[i].Moder,
			chapters[i].CreatedAt.Format(time.DateTime),
			chapters[i].UpdatedAt.Format(time.DateTime),
		)

		msg = tgbotapi.NewMessage(tgUserID, response)
		h.Bot.Send(msg)
	}

	h.Bot.Send(
		tgbotapi.NewMessage(
			tgUserID,
			`Чтобы увидеть страницы главы, используйте команду /review_chapter с указанием её названия

			Пример: /review_chapter Мертвый аккаунт`,
		),
	)
}
