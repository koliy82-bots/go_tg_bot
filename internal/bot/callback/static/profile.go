package static

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/utils/date"
	"math/rand"
	"time"
)

const (
	GrowKidData = "grow_kid"
	CasinoData  = "casino"
	HamsterData = "hamster"
)

type Opts struct {
	fx.In
	Log   *zap.Logger
	Braks brak.Repository
	Users user.Repository
	Cm    *callback.CallbacksManager
	Bot   *telego.Bot
}

func ProfileCallbacks(opts Opts) {
	opts.Cm.StaticCallback(CasinoData, func(query telego.CallbackQuery) {
		b, err := opts.Braks.FindByUserID(query.From.ID)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if date.HasUpdated(b.LastCasinoPlay) {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Играть в казино можно раз в сутки.",
				ShowAlert:       true,
			})
			return
		}

		score := rand.Intn(200) - 100
		err = opts.Braks.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{"score": score},
				"$set": bson.M{"last_casino_play": time.Now()},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}
		text := ""
		switch {
		case score > 0:
			text = fmt.Sprintf("You win %d!", score)
		case score < 0:
			text = fmt.Sprintf("You lose %d!", score)
		default:
			text = "You don't win or lose."
		}
		_, _ = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(query.Message.GetChat().ID),
			Text:   text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(GrowKidData, func(query telego.CallbackQuery) {
		b, err := opts.Braks.FindByUserID(query.From.ID)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для кормления ребёнка необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if b.BabyUserID == nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для кормления ребёнка его необходимо родить.",
				ShowAlert:       true,
			})
			return
		}

		if date.HasUpdated(b.LastGrowKid) {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Кормить ребёнка можно раз в сутки.",
				ShowAlert:       true,
			})
			return
		}

		score := rand.Intn(30) + 20
		err = opts.Braks.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{"score": score},
				"$set": bson.M{"last_grow_kid": time.Now()},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}
		text := ""
		switch {
		case score > 0:
			text = fmt.Sprintf("You win %d!", score)
		case score < 0:
			text = fmt.Sprintf("You lose %d!", score)
		default:
			text = "You don't win or lose."
		}
		_, _ = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(query.Message.GetChat().ID),
			Text:   text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(HamsterData, func(query telego.CallbackQuery) {
		b, err := opts.Braks.FindByUserID(query.From.ID)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if !date.HasUpdated(b.LastHamsterUpdate) {
			err = opts.Braks.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$inc": bson.M{"score": 1},
					"$set": bson.M{
						"tap_count":           49,
						"last_hamster_update": time.Now(),
					},
				},
			)
		} else if b.TapCount == 0 {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Хомяк устал, он разрешит себя тапать завтра.",
				ShowAlert:       true,
			})
			return
		} else {
			err = opts.Braks.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$inc": bson.M{"score": 1, "tap_count": -1},
				},
			)
		}

		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}

		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "Успешный тап по хомяку",
		})
		return

	})
}
