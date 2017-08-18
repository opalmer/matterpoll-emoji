package poll

import (
	"fmt"
	"regexp"
	"strings"
)

// Request wraps up all information needed to answer a poll request
type Request struct {
	TeamId    string
	ChannelID string
	Token     string
	Message   string
	Emojis    []string
}

const (
	back_tick          = "`"
	Error_wrong_format = `Wrong message format. Try this instead: ` + back_tick + `/poll \"What do you gys wanna grab for lunch?\" :pizza: :sushi:` + back_tick
)

// NewRequest validates the data in map and wraps it into a Request struct
func NewRequest(u map[string][]string) (*Request, error) {
	p := &Request{}
	for key, values := range u {
		switch key {
		case "team_id":
			if p.TeamId = values[0]; len(p.TeamId) == 0 {
				return nil, fmt.Errorf("Unexpected Error: TeamID in request is empty.")
			}
		case "channel_id":
			if p.ChannelID = values[0]; len(p.ChannelID) == 0 {
				return nil, fmt.Errorf("Unexpected Error: ChannelID in request is empty.")
			}
		case "token":
			if p.Token = values[0]; len(p.Token) == 0 {
				return nil, fmt.Errorf("Unexpected Error: Token in request is empty.")
			}
		case "text":
			var err error
			p.Message, p.Emojis, err = parseText(values[0])
			if err != nil {
				return nil, err
			}
		}
	}
	return p, nil
}

func parseText(text string) (string, []string, error) {
	var re *(regexp.Regexp)
	switch text[0] {
	case '`':
		re = regexp.MustCompile("`([^`]+)`(.+)")
	case '\'':
		re = regexp.MustCompile("'([^']+)'(.+)")
	case '"':
		re = regexp.MustCompile("\"([^\"]+)\"(.+)")
	default:
		return "", nil, fmt.Errorf(Error_wrong_format)
	}
	e := re.FindStringSubmatch(text)
	if len(e) != 3 {
		return "", nil, fmt.Errorf(Error_wrong_format)
	}
	var emojis []string
	for _, v := range strings.Split(e[2], " ") {
		if len(v) == 0 {
			continue
		}
		if len(v) < 3 || !strings.HasPrefix(v, ":") || !strings.HasSuffix(v, ":") {
			return "", nil, fmt.Errorf(Error_wrong_format, v)
		}
		emojis = append(emojis, v[1:len(v)-1])
	}
	return e[1], emojis, nil
}
