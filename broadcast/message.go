package broadcast

import "regexp"

const IRC_SENDER_NUMBER uint32 = 999999

type StatusType int

const (
	MESSAGE StatusType = iota
	JOIN    StatusType = iota
	PART    StatusType = iota
)

type Message struct {
	BridgeId     string
	SenderNumber uint32
	SenderName   string
	Text         string
	IsAction     bool
	Status       StatusType
}

func (message Message) Message() string {
	name := ClearName(message.SenderName)
	text := message.SenderName + ": " + message.Text

	switch message.Status {
	case JOIN:
		text = "<- " + name
	case PART:
		text = name + " ->"
	}

	return text
}

func ClearName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9 _\-]`)
	return string(re.ReplaceAllString(name, ""))
}
