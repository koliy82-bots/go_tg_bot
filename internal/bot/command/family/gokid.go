package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"time"
)

type goKidCmd struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (c goKidCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	reply := update.Message.ReplyToMessage

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
	}

	if reply == nil {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ответь на любое сообщение ребёнка.", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	b, _ := c.brakRepo.FindByUserID(from.ID)

	if b == nil {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	if b.BabyUserID != nil {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, в вашем браке уже рождён ребёнок.", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	tUser := reply.From

	if tUser.ID == from.ID || tUser.ID == b.FirstUserID || tUser.ID == b.SecondUserID {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ты не можешь стать своим же ребёнком или родить партнёра.", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	if tUser.IsBot {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, бот не может быть ребёнком.", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	kidBrakCount, _ := c.brakRepo.Count(bson.M{"baby_user_id": tUser.ID})
	if kidBrakCount != 0 {
		_, err := bot.SendMessage(params.WithDisableNotification().WithText(
			fmt.Sprintf("%s уже родился у кого-то в браке. 😥", html.UserMention(tUser))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	if time.Now().Unix() < b.CreateDate.Add(7*24*time.Hour).Unix() {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, для рождения ребёнка вы должны быть женаты неделю. ⌚", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	sUser, _ := c.userRepo.FindByID(b.PartnerID(from.ID))

	if sUser == nil {
		_, err := bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ваш партнёр не найден. 😥", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	yesCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Родиться! 🤱🏻",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := c.brakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{"$set": bson.D{
					{"baby_user_id", tUser.ID},
					{"baby_create_date", time.Now()},
				}},
			)
			if err != nil {
				c.log.Sugar().Error(err)
				return
			}
			_, err = bot.SendMessage(params.
				WithText(fmt.Sprintf("Внимание! ⚠️\n%s родился у %s и %s. 🤱",
					html.UserMention(tUser), html.UserMention(from), html.ModelMention(sUser))).
				WithReplyMarkup(nil),
			)
			if err != nil {
				c.log.Sugar().Error(err)
			}
		},
	})

	noCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Выкидыш! 😶‍🌫️",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(params.
				WithText(fmt.Sprintf("%s отказался появляться на этот свет. 💀", html.UserMention(tUser))).
				WithReplyMarkup(nil),
			)
			if err != nil {
				c.log.Sugar().Error(err)
			}
		},
	})

	_, err := bot.SendMessage(params.
		WithText(fmt.Sprintf("%s, тебе предложили родиться в семье %s и %s. 🏠",
			html.UserMention(tUser), html.UserMention(from), html.ModelMention(sUser))).
		WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(yesCallback.Inline(), noCallback.Inline()),
		)),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}

}
