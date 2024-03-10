package main

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Pong!",
	})

}

//Struct to represent the user
type User struct {
	User_id     uint   `gorm:"primaryKey"`
	Username    string `json:"username"`
	Passwd_hash string `json:"passwd_hash"`
	Created_at  time.Time
	Admin       bool
}

type Message_Obj struct {
	Sender_username   string
	Receiver_username string `json:"receiver_username"`
	Content           string `json:"content"`
}
type Convo struct {
	Convo_id     int `gorm:"primaryKey"`
	Participants []int
}

// This will trigger before user is inserted into the DB
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.Created_at = time.Now()
	return nil
}

func hash_and_salt(passwd string) (string, error) {
	passwd_hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwd_hash), nil
}

func compare_passwords(passwd string, passwd_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwd_hash), []byte(passwd))
	if err != nil {
		return false
	}
	return true
}

func register(db *gorm.DB, c *gin.Context) {
	body := User{}
	c.BindJSON(&body)
	passwd := body.Passwd_hash
	passwd_hash, err := hash_and_salt(passwd)
	body.Passwd_hash = passwd_hash
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user. Internal error."})
		return
	}

	err = put_user(db, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user. Username already exists."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User created."})

}

func login(db *gorm.DB, c *gin.Context) {
	body := User{}
	c.BindJSON(&body)
	user, err := get_user(db, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login."})
		return
	}

	if !compare_passwords(body.Passwd_hash, user.Passwd_hash) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login. Username or password incorrect."})
		return
	}
	token, err := createToken(user.Username, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User logged in.", "username": user.Username, "token": token})

}

func logged_in(c *gin.Context) (bool, jwt.MapClaims) {
	token := c.Request.Header.Get("Authorization")
	token = token[7:]
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized."})
		return false, nil
	}
	verified, claims := verifyToken(token)
	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized."})
		return false, nil
	}
	return true, claims
}

func verifyToken(token string) (bool, jwt.MapClaims) {
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("nav"), nil
	})
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok && jwtToken.Valid {
		return true, claims
	}
	return false, nil

}

func createToken(username string, exp int) (string, error) {
	// exp is in hours
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":  username,
		"exp": time.Now().Add(time.Hour * time.Duration(exp)).Unix(),
	})
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte("nav"))
	return tokenString, err
}

func sendMessage(db *gorm.DB, c *gin.Context) {
	allowed, claims := logged_in(c)
	if !allowed {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized."})
		return
	}
	if claims == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message."})
		return
	}
	msg_obj := Message_Obj{}
	c.BindJSON(&msg_obj)
	msg_obj.Sender_username = claims["ID"].(string)

	//fmt.Println(msg_obj.Sender_username, msg_obj.Receiver_username, msg_obj.Content)
	user1, err1 := get_user_by_username(db, msg_obj.Sender_username)
	user2, err2 := get_user_by_username(db, msg_obj.Receiver_username)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message."})
		return
	}
	convo_id := get_convo(db, user1.User_id, user2.User_id)
	err := put_message(db, msg_obj.Content, convo_id, int(user1.User_id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Message sent."})

}

func getLastMessage(db *gorm.DB, c *gin.Context) {
	allowed, claims := logged_in(c)
	if !allowed {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message."})
		return
	}
	if claims == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message."})
		return
	}
	msg_obj := Message_Obj{}
	c.BindJSON(&msg_obj)
	msg_obj.Sender_username = claims["ID"].(string)

	user1, err1 := get_user_by_username(db, msg_obj.Sender_username)
	user2, err2 := get_user_by_username(db, msg_obj.Receiver_username)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message."})
		return
	}
	convo_id := get_convo(db, user1.User_id, user2.User_id)
	msg, err := get_last_message(db, convo_id, user1.User_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": msg})

}

func establish(hub *Hub, c *gin.Context, db *gorm.DB, receiver string) {
	allowed, claims := logged_in(c)
	if !allowed {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized."})
		return
	}
	if claims == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fatal error."})
		return
	}
	username := claims["ID"].(string)

	username_id, err := get_user_by_username(db, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fatal error."})
		return
	}

	receiver_id, err := get_user_by_username(db, receiver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fatal error."})
		return
	}


	ServeWS(c, hub, db, username_id.User_id, receiver_id.User_id)
}


func main() {
	db, err := connect_db() 
	if err != nil {
		return
	}

	router := gin.Default()
	hub := NewHub(db)
	go hub.Run()

	router.GET("/ping", ping)
	router.POST("/register", func(c *gin.Context) {
		db, err := connect_db()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}

		register(db, c)
	})
	router.POST("/login", func(c *gin.Context) {
		db, err := connect_db() 
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}

		login(db, c) 
	})
	router.GET("/verify", func(c *gin.Context) {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}

		logged_in(c)
	})
	router.POST("/sm", func(c *gin.Context) {
		db, err := connect_db() 
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}

		sendMessage(db, c)
	})
	router.GET("/gm", func(c *gin.Context) {
		db, err := connect_db() 
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}

		getLastMessage(db, c)
	})
	router.GET("/upgrade/:receiver", func(c *gin.Context) {
		db, err := connect_db()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
			return
		}
		receiver := c.Param("receiver")

		establish(hub, c, db, receiver)
	})

	router.Run()

}

