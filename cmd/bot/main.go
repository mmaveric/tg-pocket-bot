package main

import (
	"asharipov/tg-pocket-bot/internal/db"
	"asharipov/tg-pocket-bot/internal/tg"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/genjidb/genji"
	"github.com/genjidb/genji/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var TgBotAPIToken = os.Getenv("TG_APITOKEN")

func main() {
	log.SetFlags(log.Lshortfile)
	store, err := db.NewDB("mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	bot, err := tg.NewBot(tg.BotConfig{
		ApiToken:       TgBotAPIToken,
		Debug:          false,
		MessageHandler: &DBSaveHandler{db: store},
	})
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP,
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wg.Add(1)
	go bot.Run(ctx, &wg)
	log.Println("bot started")

	wg.Wait()
	log.Println("stopped")
}

type DBSaveHandler struct {
	db *genji.DB
}

func (h *DBSaveHandler) HandleMessage(msg *tgbotapi.Message) (response string,
	reply bool) {
	if msg.IsCommand() {
		switch msg.Command() {
		case "get":
			return h.getMessagesByUsername(msg.From.UserName)
		case "remove":
			return h.deleteMessagesByUsername(msg.From.UserName), false
		case "help":
			return helpMessage, false
		case "start":
			return "", false
		default:
			return "unknown command: " + msg.Command(), false
		}

	}
	err := h.saveMessage(msg)
	if err != nil {
		return "error: " + err.Error(), false
	}
	return "saved: " + msg.Text, false
}

const helpMessage = `
/get - retrieve saved messages
/delete - delete all saved messages
/help - get this text
`

func (h *DBSaveHandler) deleteMessagesByUsername(username string) string {
	err := h.db.Exec("delete from messages where username = ?", username)
	if err != nil {
		return err.Error()
	}
	return "messages removed"
}
func (h *DBSaveHandler) getMessagesByUsername(username string) (string, bool) {
	stream, err := h.db.Query("select message from messages where username"+
		"=?", username)
	if err != nil {
		return "error: " + err.Error(), false
	}
	defer stream.Close()
	result := make([]string, 0)
	stream.Iterate(func(d types.Document) error {
		val, _ := d.GetByField("message")
		result = append(result, val.V().(string))
		return nil
	})
	return strings.Join(result, "\n"), false
}
func (h *DBSaveHandler) saveMessage(msg *tgbotapi.Message) error {
	usr := msg.From

	message := Message{
		Username:  usr.UserName,
		FirstName: usr.FirstName,
		LastName:  usr.LastName,
		Message:   msg.Text,
	}
	return h.db.Exec("insert into messages values ?;", &message)
}

type Message struct {
	FirstName string
	LastName  string
	Username  string
	Message   string
}
