package messages

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DB *sqlx.DB
)

//投稿
type PostMessageRequestBody struct {
	Text string `json:"text" from:"text"`
}

//投稿の取得
//一つの投稿
type GetMessageBody struct {
	ID       int      `json:"id" db:"id"`
	UserID   string   `json:"user_id" db:"user_id"`
	Text     string   `json:"text" db:"text"`
	PostTime string   `json:"post_time" db:"post_time"`
	FavUsers []string `json:"fav_users"`
}
type GetMessagesBody []GetMessageBody

func PostMessageHandler(c echo.Context) error {
	userID := c.Get("userID").(string)
	var req PostMessageRequestBody

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("Bad Request: %v", err))
	}

	time := time.Now()
	_, err = DB.Exec("INSERT INTO messages (user_id, text, post_time) VALUES (?, ?, ?)", userID, req.Text, time)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Db error: %v", err))
	}

	return c.NoContent(http.StatusOK)

}

func PutMessageFavHandler(c echo.Context) error {
	userID := c.Get("userID").(string)
	messageID := c.Param("id")

	var count int
	err := DB.Get(&count, "SELECT COUNT(*) FROM favolates WHERE message_id=? AND user_id=?", messageID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}
	if count > 0 {
		return c.NoContent(http.StatusOK)
	}

	_, err = DB.Exec("INSERT INTO favolates (message_id, user_id) VALUES (?, ?)", messageID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

func DeleteMessageFavHandler(c echo.Context) error {
	userID := c.Get("userID").(string)
	messageID := c.Param("id")

	_, err := DB.Exec("DELETE FROM favolates WHERE user_id=? AND message_id=?", userID, messageID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

func favUsers(messageID int) ([]string, error) {
	var userIDs []string
	err := DB.Select(&userIDs, "SELECT user_id FROM favolates WHERE message_id=?", messageID)

	if len(userIDs) == 0 {
		return []string{}, err
	}

	return userIDs, nil
}

func GetMassagesHandler(c echo.Context) error {
	var messages GetMessagesBody

	err := DB.Select(&messages, "SELECT id, user_id, text, post_time FROM messages ORDER BY id DESC")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	//favしたユーザーを取得
	for i := 0; i < len(messages); i++ {
		users, err := favUsers(messages[i].ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Internal Server Error: %v", err))
		}
		messages[i].FavUsers = users
	}
	return c.JSON(http.StatusOK, messages)
}

func GetSingleMassageHandler(c echo.Context) error {
	var message GetMessageBody
	id := c.Param("id")

	err := DB.Get(&message, "SELECT id, user_id, text, post_time FROM messages WHERE id=?", id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("db error: %v", err))
	}

	users, err := favUsers(message.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Internal Server Error: %v", err))
	}
	message.FavUsers = users

	return c.JSON(http.StatusOK, message)
}
