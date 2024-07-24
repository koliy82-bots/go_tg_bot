package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"time"
)

type goFamily struct {
	cm    *callback.CallbacksManager
	braks brak.Repository
	log   *zap.Logger
}

func (g goFamily) Handle(bot *telego.Bot, update telego.Update) {
	fUser := update.Message.From
	reply := update.Message.ReplyToMessage

	if reply == nil {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, ответь на любое сообщение партнёра. 😘💬", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	tUser := reply.From
	if tUser.ID == fUser.ID {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, брак с собой нельзя, придётся искать пару. 😥", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	if tUser.IsBot {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, бота не трогай. 👿", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	fbrak, _ := g.braks.FindByUserID(fUser.ID)

	if fbrak != nil {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, у вас уже есть брак! 💍", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	tbrak, _ := g.braks.FindByUserID(tUser.ID)

	if tbrak != nil {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, у вашего партнёра уже есть брак! 💍", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	yesCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да!❤️‍🔥",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			g.braks.Insert(&brak.Brak{
				FirstUserID:  fUser.ID,
				SecondUserID: tUser.ID,
				CreateDate:   time.Now(),
				Score:        0,
			})
			_, err := bot.SendMessage(tu.Messagef(
				telego.ChatID{ID: query.Message.GetChat().ID},
				"Hello %s!", query.From.FirstName,
			))
			if err != nil {
				g.log.Sugar().Error(err)
				return
			}
		},
	})

	noCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:      "Нет!💔",
		CtxType:    callback.ChooseOne,
		OwnerIDs:   []int64{tUser.ID},
		Time:       time.Duration(60) * time.Minute,
		AnswerText: "Отказ 🖤",
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(&telego.SendMessageParams{
				ChatID: tu.ID(update.Message.Chat.ID),
				Text:   "Отказ 🖤",
				ReplyParameters: &telego.ReplyParameters{
					MessageID: query.Message.GetMessageID(),
				},
			})
			if err != nil {
				g.log.Sugar().Error(err)
				return
			}
		},
	})

	from := update.Message.From
	_, _ = bot.SendMessage(
		tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"💍 @%s, минуточку внимания.\n"+
				"💖 @%s сделал вам предложение руки и сердца.",
			from.Username,
			from.Username,
		).WithReplyMarkup(
			tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					yesCallback.Inline(),
					noCallback.Inline(),
				),
			),
		))
}
