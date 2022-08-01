package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"

	_ "github.com/go-sql-driver/mysql"

	"github.com/joho/godotenv"
)

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

const SQL_DRIVER = "mysql"
const DB_CREDENTIAL = "root@/vkbot"

type user struct {
	id      int
	user_id int
	vote    string
	status  string
}

// Список функций которые вадают константы
func getSecretPhrases() map[string]string {
	return map[string]string{
		"fanfic": "фанфик",
		"origin": "ориджин",
	}
}

func getWellcomePhrases() map[string]string {
	return map[string]string{
		"fanfic": "Привет проголосуйте за фанфик, вот список юзеров:",
		"origin": "Привет проголосуйте за ориджин, вот список юзеров:",
	}
}

func getPhrasesForDoneUsers() map[string]string {
	return map[string]string{
		"fanfic": "Вы уже голосовали за фанфик, спасибо!!!",
		"origin": "Вы уже голосовали за ориджин, спасибо!!!",
	}
}

func getAnswerConditions() map[string]map[string]int {
	dataCondition := make(map[string]map[string]int)

	dataCondition["fanfic"] = map[string]int{}
	dataCondition["fanfic"]["min"] = 1
	dataCondition["fanfic"]["max"] = 10
	dataCondition["origin"] = map[string]int{}
	dataCondition["origin"]["min"] = 11
	dataCondition["origin"]["max"] = 15

	return dataCondition
}

//
func main() {

	vk := api.NewVK(os.Getenv("VK_ACCESS_TOKEN"))

	// get information about the group
	group, err := vk.GroupsGetByID(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Initializing Long Poll
	lp, err := longpoll.NewLongPoll(vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}

	// New message event
	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)

		userId := obj.Message.PeerID
		userMessage := obj.Message.Text
		currentUserVote := isUserWaiting(userId) // Получаем VOTE для которго ждем ответа от пользователя либо пустую строку если не чего от него не ждем

		if currentUserVote == "" { //Пользователь не должен вводить никакие сообщения
			//Проверяем, а не ключевое ли слово ввел пользоватеь
			vote := isSecretPhrase(userMessage) // Возвращаем Vote для данной ключевой фразы либо пустую строку если это просто какое то левое слово
			if vote != "" {
				if getUserStatus(userId, vote) == "done" {
					//Если по текущему голосованию пользователь уже голосовал, то  сообщаем об этом пользователю и выходим
					sendVkMessage(userId, getPhrasesForDoneUsers()[vote], vk)
					return
				}
				// Отправляем приветственную фразу
				sendVkMessage(userId, getWellcomePhrases()[vote], vk)
				//Помечаем что, ожидаем ответа от данного юзера по Vote
				setUserStatus(userId, vote, "waiting")
			}

		} else {
			userAnswer, err := strconv.Atoi(userMessage)
			if err == nil && userAnswer >= getAnswerConditions()[currentUserVote]["min"] && userAnswer <= getAnswerConditions()[currentUserVote]["max"] {
				//добовляем ответ юзера по данному голосованию в таблицу
				addUserAnswer(userId, currentUserVote, userAnswer)
				//обновляем статус юзера, что он успешно проголосовал
				updateUserStatus(userId, currentUserVote, "done")

				sendVkMessage(userId, "Спасибо за ваш голос", vk)
			} else {
				sendVkMessage(userId, "Вы должна ввести корректную цифру", vk)
			}
			// смотри какие для данного currentUserVote есть правила

		}

	})

	// Run Bots Long Poll
	log.Println("Start Long Poll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}

}
