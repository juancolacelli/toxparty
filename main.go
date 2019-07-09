package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
	"time"
	"tox_party/broadcast"
	"tox_party/irc"
	"tox_party/tox"
)

type Config struct {
	Tox *tox.Tox
	Irc []*irc.Irc
}

var messages = make(chan broadcast.Message)
var namesChanged = make(chan bool)
var config Config

func main() {
	_config, err := loadConfig()

	if err == nil {
		config = _config

		defer close(messages)
		defer close(namesChanged)

		// Update status and topic
		go func() {
			shutdown := false

			for !shutdown {
				if <-namesChanged {
					var names []string
					names = append(names, appendNames(config.Tox.Id, config.Tox.FriendsNames()))
					for _, irc := range config.Irc {
						names = append(names, appendNames(irc.Id, irc.ChannelNames()))
					}

					sort.Strings(names)

					globalNames := strings.Join(names, " - ")
					for _, irc := range config.Irc {
						irc.SetGlobalNames(globalNames)
					}

					config.Tox.SetGlobalNames(globalNames)
				}
			}
		}()

		// Broadcast
		go func() {
			shutdown := false

			for !shutdown {
				message := <-messages

				log.Println("Broadcasting message:", message)
				if message.Status == broadcast.MESSAGE {
					for _, irc := range config.Irc {
						if irc.Id != message.BridgeId {
							irc.Send(message)
						}
					}

					config.Tox.Send(message)
				}
			}
		}()

		for _, irc := range config.Irc {
			go irc.Start(messages, namesChanged)
			time.Sleep(3 * time.Second)
		}

		config.Tox.Start(messages, namesChanged)
	} else {
		log.Println("Error loading config file:", err)
	}
}

func appendNames(bridge string, names []string) string {
	result := bridge + ": "

	if bridge != "" {
		if len(names) > 0 {
			var resultNames []string
			for _, name := range names {
				name = broadcast.ClearName(name)
				if name != "" {
					resultNames = append(resultNames, name)
				}
			}

			if len(resultNames) > 0 {
				sort.Strings(resultNames)
				result += strings.Join(resultNames, ", ")
			}
		}
	}

	return result
}

func loadConfig() (Config, error) {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)

	return config, err
}
