package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"fmt"

	"github.com/boltdb/bolt"
	"strconv"
	"encoding/json"
)

const BOT_TOKEN string = "332823200:AAHEEUxiVc3EW4J1oIbPwgiPjzypp8HNZVs"
const ANTHONY int64 = 319810559

var tasks map[string]int

type Task struct {
	Name string
}
type TaskList struct {
	ID             int
	List           []Task
	CreationIsDone bool
}

type TODO struct {
	ID     int
	UserID int
	Tasks  map[string]int
}



func main() {
	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	tasks = map[string]int{"Play": 0, "English": 0}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	db, err := bolt.Open("bolt.db", 0600, nil)
	if err != nil {
		log.Fatalf("Could not open bolt db: %v", err)
	}
	defer db.Close()

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

		userID := update.Message.From.ID
		switch update.Message.Text {
		case "/start":
			db.Update(func (tx *bolt.Tx) error {
				bkt, err := tx.CreateBucketIfNotExists([]byte("Users"))
				if err != nil {
					log.Printf("Could not create bucket: %v", err)
					return err
				}
				err = bkt.Put([]byte(strconv.Itoa(userID)), []byte("true"))
				if err != nil {
					log.Printf("Could not PUT: %v", err)
				}
				return err
			})

			msg := tgbotapi.NewMessage(int64(userID), "Добавьте задачи")
			bot.Send(msg)
		case "/create":
			db.Update(func (tx *bolt.Tx) error {
				bkt, err := tx.CreateBucketIfNotExists([]byte("UsersTaskLists"))
				if err != nil {
					log.Printf("Could not create bucket: %v", err)
					return err
				}
				id, _ := bkt.NextSequence()
				tlist := TaskList{
					ID: int(id),
				}
				tlistB, err := json.Marshal(tlist)
				if err != nil {
					log.Printf("Could not marshal new tasklist: %v", err)
					return err
				}
				err = bkt.Put([]byte(strconv.Itoa(userID)), tlistB)
				if err != nil {
					log.Printf("Could not PUT: %v", err)
					return err
				}
				return err
			})
			msg := tgbotapi.NewMessage(int64(userID), "Введите задачу 1")
			bot.Send(msg)
		case "/done":
			db.Update(func(tx *bolt.Tx) error {
				bkt := tx.Bucket([]byte("UsersTaskLists"))
				if bkt != nil {
					tlistB := bkt.Get([]byte(strconv.Itoa(userID)))
					if tlistB != nil {
						tlist := TaskList{}
						err := json.Unmarshal(tlistB, &tlist)
						if err != nil {
							log.Printf("Could not unmarhal tasklist: %v", err)
							return err
						}
						tlist.CreationIsDone = true
						tlistB, err = json.Marshal(tlist)
						if err != nil {
							log.Printf("Could not marshal: %v")
							return err
						}
						bkt.Put([]byte(strconv.Itoa(userID)), tlistB)
						msg := tgbotapi.NewMessage(int64(userID), "Ваш список задач сохранен. /tasklist")
						bot.Send(msg)
						return nil
					}
				}
				return nil
			})
		case "/tasklist":
			db.View(func(tx *bolt.Tx) error {
				bkt := tx.Bucket([]byte("UsersTaskLists"))
				if bkt != nil {
					tlistB := bkt.Get([]byte(strconv.Itoa(userID)))
					var count int
					if tlistB != nil {
						tlist := TaskList{}
						json.Unmarshal(tlistB, tlist)
						count = len(tlist.List)
					}

					msg := tgbotapi.NewMessage(int64(userID), "У вас всего " + strconv.Itoa(count) + " задач.")
					bot.Send(msg)
				}
				return nil
			})
		default:
			db.Update(func(tx *bolt.Tx) error {
				bkt := tx.Bucket([]byte("UsersTaskLists"))
				if bkt != nil {
					tlistB := bkt.Get([]byte(strconv.Itoa(userID)))
					// determine Task creation process
					if tlistB != nil {
						tlist := TaskList{}
						err := json.Unmarshal(tlistB, &tlist)
						if err != nil {
							log.Printf("Could not unmarhal tasklist: %v", err)
							return err
						}
						if !tlist.CreationIsDone {
							task := Task{
								Name: update.Message.Text,
							}
							tlist.List = append(tlist.List, task)
							tlistB, err = json.Marshal(tlist)
							if err != nil {
								log.Printf("Could not marshal: %v")
								return err
							}
							bkt.Put([]byte(strconv.Itoa(userID)), tlistB)
							msg := tgbotapi.NewMessage(int64(userID), "Введите следующую задачу либо завершите /done")
							bot.Send(msg)
							return nil
						}
					}
					msg := tgbotapi.NewMessage(int64(userID), "Доступные команды:\n/tasks   - Список дел")
					bot.Send(msg)
				}
				return nil
			})
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
