package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/utils/html"
	"time"
)

type goKid struct {
	cm    *callback.CallbacksManager
	braks brak.Repository
	users user.Repository
	log   *zap.Logger
}

func (g goKid) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	reply := update.Message.ReplyToMessage

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: update.Message.GetMessageID(),
		},
	}

	if reply == nil {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ответь на любое сообщение ребёнка.", html.UserMention(from))),
		)
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	b, _ := g.braks.FindByUserID(from.ID)

	if b == nil {
		_, _ = bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		return
	}

	if b.BabyUserID != nil {
		_, _ = bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, в вашем браке уже рождён ребёнок.", html.UserMention(from))),
		)
		return
	}

	tUser := reply.From

	if tUser.ID == from.ID || tUser.ID == b.FirstUserID || tUser.ID == b.SecondUserID {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ты не можешь стать своим же ребёнком или родить партнёра.", html.UserMention(from))),
		)
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	if tUser.IsBot {
		_, _ = bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, бот не может быть ребёнком.", html.UserMention(from))),
		)
		return
	}

	kidBrak, _ := g.braks.FindByKidID(tUser.ID)
	if kidBrak != nil {
		_, _ = bot.SendMessage(params.WithDisableNotification().WithText(
			fmt.Sprintf("%s уже родился у кого-то в браке. 😥", html.UserMention(tUser))),
		)
		return
	}

	//if time.Now().Unix() < b.CreateDate.Add(7*24*time.Hour).Unix() {
	//	_, _ = bot.SendMessage(params.WithText(
	//		fmt.Sprintf("%s, для рождения ребёнка вы должны быть женаты неделю. ⌚", html.UserMention(from))),
	//	)
	//	return
	//}

	sUser, _ := g.users.FindByID(b.PartnerID(from.ID))

	if sUser == nil {
		return
	}

	yesCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Родиться! 🤱🏻",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			//baby_create_date := time.Now()
			err := g.braks.Update(
				bson.M{"_id": b.OID},
				bson.M{"$set": bson.D{
					{"baby_user_id", tUser.ID},
					{"baby_create_date", time.Now()},
				}},
			)
			if err != nil {
				g.log.Sugar().Error(err)
				return
			}
			_, _ = bot.SendMessage(params.
				WithText(fmt.Sprintf("Внимание! ⚠️\n%s родился у %s и %s. 🧑🏽‍👩🏽‍🧒🏿",
					html.UserMention(tUser), html.UserMention(from), sUser.Mention())).
				WithReplyMarkup(nil),
			)
		},
	})

	noCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Выкидыш! 😶‍🌫️",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, _ = bot.SendMessage(params.
				WithText(fmt.Sprintf("%s отказался появляться на этот свет. 💀", html.UserMention(tUser))).
				WithReplyMarkup(nil),
			)
		},
	})

	_, _ = bot.SendMessage(params.
		WithText(fmt.Sprintf("%s, тебе предложили родиться в семье %s и %s. 🏠",
			html.UserMention(tUser), html.UserMention(from), sUser.Mention())).
		WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(yesCallback.Inline(), noCallback.Inline()),
		)),
	)

}
