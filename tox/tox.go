package tox

import (
	toxLib "github.com/TokTok/go-toxcore-c"
	"io/ioutil"
	"log"
	"time"
	"tox_party/broadcast"
)

var friends = make(map[uint32]Friend)
var friendsNames []string

type Tox struct {
	client      *toxLib.Tox
	globalNames string

	Id     string
	Name   string
	Status string
	File   string
}

func (tox *Tox) Start(broadcasts chan broadcast.Message, namesChanged chan bool) {
	if tox.Id == "" {
		tox.Id = "tox"
	}

	opt := toxLib.NewToxOptions()

	if toxLib.FileExist(tox.File) {
		data, err := ioutil.ReadFile(tox.File)
		if err == nil {
			opt.Savedata_data = data
			opt.Savedata_type = toxLib.SAVEDATA_TYPE_TOX_SAVE
		} else {
			log.Println(err)
		}
	}

	tox.client = toxLib.NewTox(opt)
	log.Println("Tox ID:", tox.client.SelfGetAddress())

	tox.saveToxFile()

	tox.client.CallbackSelfConnectionStatus(func(t *toxLib.Tox, status int, userData interface{}) {
		log.Println("Status:", status, userData)

		if status == 2 {
			tox.client.SelfSetName(tox.Name)
			log.Println("Name:", tox.Name)

			tox.client.SelfSetStatusMessage(tox.Status)
			log.Println("Status:", tox.Status)
		}
	}, nil)

	tox.client.CallbackFriendRequest(func(t *toxLib.Tox, friendId string, message string, userData interface{}) {
		num, err := tox.client.FriendAddNorequest(friendId)
		log.Println("Accepting friend:", num, err)

		tox.saveToxFile()
	}, nil)

	tox.client.CallbackFriendConnectionStatus(func(this *toxLib.Tox, friendNumber uint32, status int, userData interface{}) {
		friend, exists := friends[friendNumber]

		if !exists {
			friendName, _ := tox.client.FriendGetName(friendNumber)
			friendPublicKey, _ := tox.client.FriendGetPublicKey(friendNumber)
			friend = Friend{
				Number:    friendNumber,
				Name:      friendName,
				PublicKey: friendPublicKey,
			}
		}

		statusType := broadcast.JOIN

		switch status {
		case 0:
			friend.IsOnline = false
			statusType = broadcast.PART

		default:
			friend.IsOnline = true
		}

		friends[friendNumber] = friend

		tox.updateFriendsNames()
		namesChanged <- true

		log.Println("Friend status:", friend.Name, friend.IsOnline, status)

		timer := time.NewTimer(5 * time.Second)
		go func() {
			<-timer.C
			updatedFriend := friends[friend.Number]

			if updatedFriend.IsOnline == friend.IsOnline {
				broadcasts <- broadcast.Message{
					BridgeId:     tox.Id,
					SenderNumber: friend.Number,
					SenderName:   friend.Name,
					Status:       statusType,
				}
			}
		}()
	}, nil)

	tox.client.CallbackFriendName(func(this *toxLib.Tox, friendNumber uint32, newName string, userData interface{}) {
		friend, _ := friends[friendNumber]
		friend.Name = newName
		friends[friendNumber] = friend

		tox.updateFriendsNames()
		namesChanged <- true

		log.Println("Friend name:", newName)
	}, nil)

	tox.client.CallbackFriendMessage(func(t *toxLib.Tox, friendNumber uint32, message string, userData interface{}) {
		log.Println("Tox message:", message)

		sender := friends[friendNumber]

		switch message {
		case "!on":
			tox.client.FriendSendMessage(friendNumber, tox.globalNames)

		default:
			broadcasts <- broadcast.Message{
				BridgeId:     tox.Id,
				SenderNumber: sender.Number,
				SenderName:   sender.Name,
				Text:         message,
			}
		}

	}, nil)

	shutdown := false

	for !shutdown {
		tox.client.IterationInterval()
		tox.client.Iterate()

		time.Sleep(1000 * 50 * time.Microsecond)
	}

	tox.client.Kill()
}

func (tox *Tox) send(message Message) {
	for _, receiver := range friends {
		// Prevent sending to sender, senderNumber cannot be null, so the IRC number will be forced
		if (receiver.Number != message.SenderNumber || message.SenderNumber == broadcast.IRC_SENDER_NUMBER) && receiver.IsOnline {
			if message.IsAction {
				tox.client.FriendSendAction(receiver.Number, message.Text)
			} else {
				tox.client.FriendSendMessage(receiver.Number, message.Text)
			}
		}
	}
}

func (tox *Tox) Send(message broadcast.Message) {
	log.Println("Sending message:", tox.Id, message)

	toxMessage := Message{
		SenderNumber: message.SenderNumber,
		SenderName:   message.SenderName,
		Text:         message.Message(),
		IsAction:     message.IsAction,
	}

	tox.send(toxMessage)
}

func (tox *Tox) saveToxFile() {
	err := tox.client.WriteSavedata(tox.File)
	log.Println("Tox data saved: ", err)
}

func (tox *Tox) updateFriendsNames() {
	var names []string

	for _, friend := range friends {
		if friend.IsOnline && friend.Name != "" {
			names = append(names, friend.Name)
		}
	}

	friendsNames = names
}

func (tox *Tox) FriendsNames() ([]string) {
	return friendsNames
}

func (tox *Tox) SetGlobalNames(names string) {
	tox.globalNames = names
}
