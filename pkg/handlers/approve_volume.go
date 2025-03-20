package handlers

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/Araks1255/mangacage/pkg/common/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h handler) ApproveVolume(update tgbotapi.Update) {
	tgUserID := update.Message.Chat.ID

	userID, ok := h.AllowedUsers[tgUserID]
	if !ok {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Вы не являетесь модератором или администратором"))
		return
	}

	volumeOnModerationID, err := strconv.Atoi(update.Message.CommandArguments())
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Введите id обращения тома, который хотите одобрить\n\nПример: /approve_volume 3"))
		return
	}

	var volumeOnModeration models.VolumeOnModeration
	h.DB.Raw("SELECT * FROM volumes_on_moderation WHERE id = ?", volumeOnModerationID).Scan(&volumeOnModeration)
	if volumeOnModeration.ID == 0 {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "На модерации нет тома с таким id обращения"))
		return
	}

	var volume models.Volume
	h.DB.Raw("SELECT * FROM volumes WHERE id = ?", volumeOnModeration.ExistingID).Scan(&volume)

	doesVolumeExist := volume.ID != 0

	ConvertToVolume(volumeOnModeration, &volume)
	volume.ModeratorID = sql.NullInt64{Int64: int64(userID), Valid: true}

	tx := h.DB.Begin()

	if result := tx.Save(&volume); result.Error != nil {
		tx.Rollback()
		log.Println(result.Error)
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Произошла ошибка при создании или обновлении тома"))
		return
	}

	tx.Commit()

	if doesVolumeExist {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том успешно изменён"))
	} else {
		h.Bot.Send(tgbotapi.NewMessage(tgUserID, "Том успешно создан"))
	}

	if result := h.DB.Exec("DELETE FROM volumes_on_moderation WHERE id = ?", volumeOnModeration.ID); result.Error != nil {
		log.Println(result.Error)
	}
}

func ConvertToVolume(volumeOnModeration models.VolumeOnModeration, volume *models.Volume) {
	volume.Name = volumeOnModeration.Name
	volume.Description = volumeOnModeration.Description
	volume.TitleID = volumeOnModeration.TitleID
	volume.CreatorID = volumeOnModeration.CreatorID
}
