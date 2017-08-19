package poll

import (
	"fmt"
	"github.com/mattermost/platform/model"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// RESPONSE_USERNAME is the username which will be used to post the slack command response
	RESPONSE_USERNAME = "Matterpoll"
	// RESPONSE_ICON_URL is the profile picture which will be used to post the slack command response
	RESPONSE_ICON_URL = "https://www.mattermost.org/wp-content/uploads/2016/04/icon.png"
)

var C *Conf

// Cmd handles a slash command request and sends back a response
func Cmd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// Check if Content Type is correct
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	poll, err := NewRequest(r.Form)
	validPoll := err == nil

	var response model.CommandResponse
	response.Username = RESPONSE_USERNAME
	response.IconURL = RESPONSE_ICON_URL
	if validPoll {
		response.ResponseType = model.COMMAND_RESPONSE_TYPE_IN_CHANNEL
		response.Text = poll.Message + ` #poll`
	} else {
		response.ResponseType = model.COMMAND_RESPONSE_TYPE_EPHEMERAL
		response.Text = err.Error()
	}
	io.WriteString(w, response.ToJson())
	if validPoll {
		if len(C.Token) != 0 && C.Token != poll.Token {
			log.Print("Token missmatch. Check you config.json")
			return
		}

		c := model.NewAPIv4Client(C.Host)
		user, err := login(c)
		if err != nil {
			log.Print(err)
			return
		}
		go addReaction(c, user, poll)
	}
}

func login(c *model.Client4) (*model.User, error) {
	u, apiResponse := c.Login(C.User.ID, C.User.Password)
	if apiResponse != nil && apiResponse.StatusCode != 200 {
		return nil, fmt.Errorf("Error: Login failed. API statuscode: %v", apiResponse.StatusCode)
	}
	return u, nil
}

func addReaction(c *model.Client4, user *model.User, poll *Request) {
	for try := 0; try < 5; try++ {
		// Get the last post and compare it to our message text
		result, apiResponse := c.GetPostsForChannel(poll.ChannelID, 0, 1, "")
		if apiResponse != nil && apiResponse.StatusCode != 200 {
			log.Printf("Error: Failed to fetch posts. API statuscode: %v", apiResponse.StatusCode)
			return
		}
		postID := result.Order[0]
		if result.Posts[postID].Message == poll.Message+" #poll" {
			err := reaction(c, poll.ChannelID, user.Id, postID, poll.Emojis)
			if err != nil {
				log.Print(err)
				return
			}
			return
		}
		// Try again later
		time.Sleep(100 * time.Millisecond)
	}
}

func reaction(c *model.Client4, channelID string, userID string, postID string, emojis []string) error {
	for _, e := range emojis {
		r := model.Reaction{
			UserId:    userID,
			PostId:    postID,
			EmojiName: e,
		}
		_, apiResponse := c.SaveReaction(&r)
		if apiResponse != nil && apiResponse.StatusCode != 200 {
			return fmt.Errorf("Error: Failed to save reaction. API statuscode: %v", apiResponse.StatusCode)
		}
	}
	return nil
}
