package irc

import (
	"crypto/tls"
	ircLib "github.com/thoj/go-ircevent"
	"log"
	"strings"
	"tox_party/broadcast"
)

type Irc struct {
	client       *ircLib.Connection
	channelNames []string
	globalNames  string

	Id             string
	Nick           string
	User           string
	Name           string
	Server         string
	ServerPassword string
	UseSSL         bool
	Channel        string
}

func (irc *Irc) Start(broadcasts chan broadcast.Message, namesChanged chan bool) {
	irc.client = ircLib.IRC(irc.Nick, irc.User)
	irc.client.RealName = irc.Name

	if irc.ServerPassword != "" {
		irc.client.Password = irc.ServerPassword
	}

	if irc.UseSSL {
		irc.client.UseTLS = irc.UseSSL
		irc.client.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	irc.client.AddCallback("001", func(event *ircLib.Event) {
		if irc.Id == "" {
			irc.Id = strings.Split(event.Raw, " ")[0][1:]
		}

		log.Println("Server name:", irc.Id)

		log.Println("Joining channel:", irc.Channel)
		irc.client.Join(irc.Channel)
	})

	irc.client.AddCallback("PRIVMSG", func(event *ircLib.Event) {
		log.Println("IRC message:", event.Message())

		switch event.Message() {
		case "!on":
			irc.client.Privmsg(irc.Channel, irc.globalNames)
		default:
			broadcasts <- broadcast.Message{
				BridgeId:     irc.Id,
				SenderNumber: broadcast.IRC_SENDER_NUMBER,
				SenderName:   event.Nick,
				Text:         event.Message(),
			}
		}
	})

	irc.client.AddCallback("CTCP_ACTION", func(event *ircLib.Event) {
		log.Println("IRC action:", event.Message())

		broadcasts <- broadcast.Message{
			BridgeId:     irc.Id,
			SenderNumber: broadcast.IRC_SENDER_NUMBER,
			SenderName:   event.Nick,
			Text:         event.Message(),
			IsAction:     true,
		}
	});
	irc.client.AddCallback("KICK", func(event *ircLib.Event) {
		irc.client.Join(irc.Channel)
		irc.getNames()

		broadcasts <- broadcast.Message{
			BridgeId:     irc.Id,
			SenderNumber: broadcast.IRC_SENDER_NUMBER,
			SenderName:   event.Nick,
			Status:       broadcast.PART,
		}
	})

	irc.client.AddCallback("JOIN", func(event *ircLib.Event) {
		irc.getNames()

		if irc.client.GetNick() != event.Nick {
			broadcasts <- broadcast.Message{
				BridgeId:     irc.Id,
				SenderNumber: broadcast.IRC_SENDER_NUMBER,
				SenderName:   event.Nick,
				Status:       broadcast.JOIN,
			}
		}
	})

	irc.client.AddCallback("PART", func(event *ircLib.Event) {
		irc.getNames()

		broadcasts <- broadcast.Message{
			BridgeId:     irc.Id,
			SenderNumber: broadcast.IRC_SENDER_NUMBER,
			SenderName:   event.Nick,
			Status:       broadcast.PART,
		}
	})

	irc.client.AddCallback("QUIT", func(event *ircLib.Event) {
		irc.getNames()

		broadcasts <- broadcast.Message{
			BridgeId:     irc.Id,
			SenderNumber: broadcast.IRC_SENDER_NUMBER,
			SenderName:   event.Nick,
			Status:       broadcast.PART,
		}
	})

	irc.client.AddCallback("353", func(event *ircLib.Event) {
		var names []string
		for _, name := range strings.Split(event.Message(), " ") {
			if name != irc.client.GetNick() {
				names = append(names, name)
			}
		}

		irc.channelNames = names
		namesChanged <- true
	})

	err := irc.client.Connect(irc.Server)
	if err != nil {
		log.Println("Error connecting IRC server:", err)
	}
}

func (irc *Irc) Send(message broadcast.Message) {
	if irc.client.Connected() {
		log.Println("Sending message:", irc.Id, message)
		irc.client.Privmsg(irc.Channel, message.Message())
	} else {
		log.Println("IRC not connected")
	}
}

func (irc *Irc) getNames() {
	if irc.client.Connected() {
		log.Println("Getting names")
		irc.client.SendRaw("NAMES " + irc.Channel)
	} else {
		log.Println("IRC not connected")
	}
}

func (irc *Irc) ChannelNames() ([]string) {
	return irc.channelNames
}

func (irc *Irc) SetGlobalNames(names string) {
	irc.globalNames = names
}
