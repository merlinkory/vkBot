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

var SQL_DRIVER string

var DB_CREDENTIAL string

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
		"origin": "оридж",
	}
}

func getWellcomePhrases() map[string]string {
	return map[string]string{
		"fanfic": "Привет! Список участников конкурса сетературы в номинации «фанфикшен»:\n1. https://clck.ru/sTvSP\n2. https://clck.ru/sTvSc\n3. https://clck.ru/sTvSy\n4. https://clck.ru/sTvTD\n\nЧтобы проголосовать за понравившуюся работу, пришлите ее порядковый номер (только цифры).\nПроголосовать можно лишь один раз, изменить выбор, после того как проголосуете, будет нельзя.",
		"origin": "Привет! Список участников конкурса сетературы в номинации «ориджинал фикшен»:\n1. О неподходящих сказках и неожиданных концовках семейных легенд: https://clck.ru/sSJdM\n2. Поцелуй электродрели https://clck.ru/sSJew\n3. Волшебное кушанье https://clck.ru/sSJfn\n4. Шутники https://clck.ru/sSJgV\n5. Game over https://clck.ru/sSJgx\n6. Кружавчики https://clck.ru/sSJhe\n7. Лилия демона https://clck.ru/sSJiL\n8. Вооружена и особо «опасна» https://clck.ru/sSJjL\n9. Беседа https://clck.ru/sSJkM\n10. Играй! https://clck.ru/sSJnU\n11. «Надежда», или Плавание в край жемчуга https://clck.ru/sSJoJ\n12. Дело полное изменений https://clck.ru/sSJph\n13. В этом мире смерти нет? https://clck.ru/sSJqz\n14. Дар https://clck.ru/sTvPW\n15. «Он врывается без стука и уходит, не сказав» https://clck.ru/sSJtW\n\nЧтобы проголосовать за понравившуюся работу, пришлите ее порядковый номер (только цифры).\n Проголосовать можно лишь один раз, изменить выбор, после того как проголосуете, будет нельзя.",
	}
}

func getPhrasesForDoneUsers() map[string]string {
	return map[string]string{
		"fanfic": "Привет! Ты уже проголосовал в этом конкурсе. Увы, повторно проголосовать нельзя.",
		"origin": "Привет! Ты уже проголосовал в этом конкурсе. Увы, повторно проголосовать нельзя.",
	}
}

func getAnswerConditions() map[string]map[string]int {
	dataCondition := make(map[string]map[string]int)

	dataCondition["fanfic"] = map[string]int{}
	dataCondition["fanfic"]["min"] = 1
	dataCondition["fanfic"]["max"] = 4
	dataCondition["origin"] = map[string]int{}
	dataCondition["origin"]["min"] = 1
	dataCondition["origin"]["max"] = 15

	return dataCondition
}

//
func main() {

	SQL_DRIVER = os.Getenv("SQL_DRIVER")
	DB_CREDENTIAL = os.Getenv("DB_CREDENTIAL")
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

				sendVkMessage(userId, "Спасибо, ваш голос учтен! =)", vk)
			} else {
				sendVkMessage(userId, "Привет! Ты ввел непонятную команду. Чтобы проголосовать за понравившуюся работу, пришлите ее порядковый номер (только цифры). То есть, если тебе понравилась работа под номером 1, просто напиши в ответном сообщении одну цифру: 1", vk)
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
