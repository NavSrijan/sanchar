package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connect_db() (*gorm.DB, error) {
	connectionString := "user=thewhistler dbname=test sslmode=disable"

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	return db, nil

}

func put_user(db *gorm.DB, user User) error {
	//_, err := db.Query("INSERT INTO users (username, passwd_hash) VALUES ('thewhistler', 'password1234')")
	//fmt.Println("Creating user: ", user.username)
	//result := db.Select("Username", "Passwd_hash", "Created_at").Create(&user)
	result := db.Create(&user)
	fmt.Println(result.Error)
	//fmt.Println("User created with id: ", user.user_id)
	return result.Error
}

func get_user(db *gorm.DB, user User) (User, error) {
	result := db.Where("username = ?", user.Username).First(&user)
	if result.Error != nil {
		//fmt.Println("Error: ", result.Error)
		return user, result.Error
	}
	return user, nil
}

func get_user_by_username(db *gorm.DB, username string) (User, error) {
	user := User{}
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		//fmt.Println("Error: ", result.Error)
		return user, result.Error
	}
	return user, nil
}

func run_query(db *gorm.DB, query string) {
	rows := db.Raw(query)
	_ = rows
}

func get_convo(db *gorm.DB, u1 uint, u2 uint) int {
	convo_id := -1
	result := db.Raw("SELECT convo_id FROM convos WHERE participants @> ARRAY[?, ?]::int[]", u1, u2).Scan(&convo_id)
	if result.Error != nil || convo_id == -1 {
		result = db.Raw("INSERT INTO convos (participants) VALUES(ARRAY[?, ?]::int[])", u2, u1).Scan(&convo_id)
		if result.Error != nil {
			fmt.Println("Error: ", result.Error)
		}

		result = db.Raw("SELECT convo_id FROM convos WHERE participants @> ARRAY[?, ?]::int[]", u1, u2).Scan(&convo_id)
		if result.Error != nil {
			fmt.Println("Error: ", result.Error)
		}

	}
	return convo_id
}

func put_message(db *gorm.DB, message string, convo_id int, user_id int) error {
	result := db.Exec("INSERT INTO messages (user_id, convo_id, content) VALUES (?, ?, ?)", user_id, convo_id, message)
	err := result.Error
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return err
}

func get_last_message(db *gorm.DB, convo_id int, user1_id uint) (string, error) {
	msg := ""
	result := db.Raw("SELECT content FROM messages WHERE convo_id = ? and user_id != ? ORDER BY created_at DESC LIMIT 10", convo_id, user1_id).Scan(&msg)
	if result.Error != nil {
		fmt.Println("Error: ", result.Error)
		return msg, result.Error
	}
	return msg, nil
}

