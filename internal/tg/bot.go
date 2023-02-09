package tg

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler interface {
	HandleMessage(usr *tgbotapi.Message) (response string,
		reply bool)
}

type BotConfig struct {
	ApiToken       string
	Debug          bool
	MessageHandler MessageHandler
}
type Bot struct {
	bot     *tgbotapi.BotAPI
	handler MessageHandler
	debug   bool
}

func NewBot(cfg BotConfig) (*Bot, error) {
	b := &Bot{
		debug:   cfg.Debug,
		handler: cfg.MessageHandler,
	}
	var err error
	b.bot, err = tgbotapi.NewBotAPI(cfg.ApiToken)
	if err != nil {
		return nil, err
	}
	b.bot.Debug = cfg.Debug
	return b, nil
}

func (b *Bot) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Println("bot stopped")
	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	updates := b.bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			if b.handler != nil {
				response, reply := b.handler.HandleMessage(update.Message)
				if response != "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
					msg.ReplyMarkup = kb
					if reply {
						msg.ReplyToMessageID = update.Message.MessageID
					}
					if _, err := b.bot.Send(msg); err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

var kb = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/get"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/help"),
	),
)
