package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"fmt"
)

const BOT_TOKEN string = "332823200:AAHEEUxiVc3EW4J1oIbPwgiPjzypp8HNZVs"
const ANTHONY int64 = 319810559

var tasks map[string]int


func main() {
	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	tasks = map[string]int{"Play": 0, "English": 0}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Could not get updates: %v", err)
	}
	for update := range updates {
		if update.Message == nil {
			if update.CallbackQuery != nil {
				var text string
				switch update.CallbackQuery.Data {
				case "Play":
					text = "С Витей сегодня поиграли!"
					tasks["Play"]++
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, text)
					bot.AnswerCallbackQuery(callback)
					msg := tgbotapi.NewEditMessageText(int64(update.CallbackQuery.From.ID),
														update.CallbackQuery.Message.From.ID,
														"Тратара")
					bot.Send(msg)
					bot.Send(newTasksMsg(int64(update.CallbackQuery.From.ID)))
				case "English":
					text = "Английский сделали"
					tasks["English"]++
					callback := tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, text)
					bot.AnswerCallbackQuery(callback)
				}

			} else if update.InlineQuery != nil {
				if update.InlineQuery.Query == "/tasks" {
					msg := tgbotapi.NewMessage(ANTHONY, "Нужно делать поддержку inline-mode")
					bot.Send(msg)
				}
			}
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text == "/tasks" {
			msg := newTasksMsg(int64(update.Message.From.ID))
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(int64(update.Message.From.ID), "Доступные команды:\n/tasks   - Список дел")
			bot.Send(msg)
		}
	}
}

func newTasksMsg(id int64) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(id, fmt.Sprintf("Витя %d\nАнглийский %d", tasks["Play"], tasks["English"]))
	btn1 := tgbotapi.NewInlineKeyboardButtonData("Поиграть с Витей", "Play")
	btn2 := tgbotapi.NewInlineKeyboardButtonData("Английский", "English")
	row1 := tgbotapi.NewInlineKeyboardRow(btn1, btn2)
	row2 := tgbotapi.NewInlineKeyboardRow(btn2)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row1, row2)
	msg.ReplyMarkup = keyboard
	return msg
}
