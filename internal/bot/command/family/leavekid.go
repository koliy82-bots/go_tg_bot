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
	"go_tg_bot/internal/pkg/html"
	"math/rand/v2"
	"time"
)

type leaveKid struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (e leaveKid) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
	}
	b, _ := e.brakRepo.FindByKidID(from.ID)
	if b == nil {
		_, _ = bot.SendMessage(params.WithText(
			fmt.Sprintf("%s, ты ещё не родился. ⌚", html.UserMention(from))),
		)
		return
	}

	yesCallback := e.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да.",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{from.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := e.brakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{"$set": bson.D{
					{"baby_user_id", nil},
					{"baby_create_date", nil},
				}},
			)
			if err != nil {
				e.log.Sugar().Error(err)
				return
			}

			r := rand.IntN(45)
			text := ""
			switch {
			case r >= 0 && r <= 15:
				text = fmt.Sprintf("%s шёл по дороге к детскому дому, когда вдруг на него напало стадо белых и пушистых котиков! Они носили его на лапках, подбрасывали в воздух и играли с ним в прятки. Веселая котячья армия проводила его прямо до дверей детского дома, где его ждали с радостью и открытыми объятиями.", html.UserMention(from))
			case r >= 16 && r <= 20:
				text = fmt.Sprintf("%s решил бежать от своих родителей и отправиться в детский дом, мечтая о лучшей жизни и большей заботе. Он собрал свои немногочисленные вещи и тихонько вышел из дома в темноте. По пути он столкнулся с непредвиденными преградами, но его решимость не ослабевала. Однако, в долине, которую он пытался пересечь, случился сильный ливень. Ребёнок оказался в беде и без надежды на помощь. Он лежал на земле, мокрый и испуганный, пока его силы постепенно оставляли его тело. Никто не знал о его печали и потере, и его мечты о лучшей жизни исчезли вместе с его последним дыханием.", html.UserMention(from))
			case r >= 21 && r <= 30:
				text = fmt.Sprintf("%s устав от бесконечных правил и запретов, которые накладывали на него родители, решил покинуть свой дом и отправиться в детский дом. Когда он пришёл туда, сотрудники были поражены его решимостью и сказали: \"Мы рады принять тебя с открытыми объятиями, малыш! Здесь ты найдёшь новый дом и новую семью, которая будет заботиться о тебе.\" Ребёнок улыбнулся и понял, что он сделал правильный выбор, и вместе с новыми друзьями начал своё новое приключение в детском доме.", html.UserMention(from))
			case r >= 31 && r <= 40:
				text = fmt.Sprintf("%s шёл на концерт Моргенштерна, он был настолько взбудоражен и восторжен, что его энергия переполняла его самого. Он подпрыгивал и танцевал на своем пути, неся с собой невероятное веселье. Но внезапно его энтузиазм перешел в пределы возможного, и он начал сверкать ярким светом, превращаясь в маленькую звезду. Все, кто видел это чудо, восхищенно замирали, понимая, что это было что-то особенное. И хотя его приключение закончилось раньше времени, ребёнок оставил память о своей неподражаемой энергии и радости в сердцах всех, кто видел его.", html.UserMention(from))
			case r >= 41 && r <= 45:
				text = fmt.Sprintf("%s, шёл к детскому дому, неся с собой свою крутую механику из Dota 2. Внезапно, перед ним выскочил сильный вражеский герой, и ребёнок сразу же активировал свои навыки. Он прыгнул в воздух, развернулся и нанёс мощный удар, отправляя врага в космос. Прохожие ошарашено остановились, а ребёнок уверенно продолжил свой путь, собирая аплодисменты и восхищённые взгляды.", html.UserMention(from))
			default:
				text = "Что-то пошло не так..."
			}

			_, _ = bot.SendMessage(params.
				WithText(text).
				WithReplyMarkup(nil),
			)
		},
	})

	_, _ = bot.SendMessage(params.
		WithText(fmt.Sprintf("%s, ты уверен, что хочешь покинуть свою семью? 🏠", html.UserMention(from))).
		WithReplyMarkup(tu.InlineKeyboard(tu.InlineKeyboardRow(yesCallback.Inline()))),
	)
}
