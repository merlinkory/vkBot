package main

import (
	"database/sql"
	"fmt"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
)

func isSecretPhrase(phrase string) string {
	phrases := getSecretPhrases()

	for k, v := range phrases {
		if v == phrase {
			return k
		}
	}
	return ""
}

// Отправляем сообщение в ВК
func sendVkMessage(id int, text string, vk *api.VK) bool {
	res := params.NewMessagesSendBuilder()
	res.Message(text)
	res.RandomID(0)
	res.PeerID(id)
	_, err := vk.MessagesSend(res.Params)

	return err == nil
}

// добовляем ответ пользователя в БД
func addUserAnswer(user_id int, vote string, answer int) {
	db, err := sql.Open(SQL_DRIVER, DB_CREDENTIAL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec("insert into answers (user_id, vote, answer) values (?, ?, ?)", user_id, vote, answer)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.LastInsertId()) // id добавленного объекта
	fmt.Println(result.RowsAffected()) // количество затронутых строк
}
func updateUserStatus(user_id int, vote string, status string) {
	db, err := sql.Open(SQL_DRIVER, DB_CREDENTIAL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec("UPDATE answer_users SET status = ? where user_id = ? AND vote = ?", status, user_id, vote)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.LastInsertId()) // id добавленного объекта
	fmt.Println(result.RowsAffected()) // количество затронутых строк
}

// добовляем нового пользователя в БД
func setUserStatus(user_id int, vote string, status string) {
	db, err := sql.Open(SQL_DRIVER, DB_CREDENTIAL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec("insert into answer_users (user_id, vote, status) values (?,?,?)", user_id, vote, status)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.LastInsertId()) // id добавленного объекта
	fmt.Println(result.RowsAffected()) // количество затронутых строк
}

func isUserWaiting(user_id int) string {
	db, err := sql.Open(SQL_DRIVER, DB_CREDENTIAL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `SELECT * FROM answer_users WHERE user_id = ? AND status = 'waiting'`
	userData := user{}
	row := db.QueryRow(sqlStatement, user_id)
	err = row.Scan(&userData.id, &userData.user_id, &userData.vote, &userData.status)
	if err != nil {
		if err == sql.ErrNoRows {
			return ""
		} else {
			panic(err)
		}
	}

	return userData.vote
}

/*
 * Проверяем нужно ли от данного юзера ждать ответ на голосование и проголосовал ли он
 */
func getUserStatus(user_id int, vote string) string {

	db, err := sql.Open(SQL_DRIVER, DB_CREDENTIAL)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	sqlStatement := `SELECT * FROM answer_users WHERE user_id = ? AND vote = ?`
	userData := user{}
	row := db.QueryRow(sqlStatement, user_id, vote)
	err = row.Scan(&userData.id, &userData.user_id, &userData.vote, &userData.status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "new"
		} else {
			panic(err)
		}
	}

	return userData.status
}
