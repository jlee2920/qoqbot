package main

import (
	"fmt"
	"strconv"
	"strings"

	// _ "github.com/jinzhu/gorm/dialects/mssql"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jlee2920/qoqbot.git/config"
	youtube "google.golang.org/api/youtube/v3"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/gempir/go-twitch-irc"
	_ "github.com/lib/pq"
)

// Starts a new Twitch IRC client that listens for messages sent
func startTwitchIRC(service *youtube.Service) {
	fmt.Print("Starting server!\n")
	// Instantiate a new client
	client := twitch.NewClient(config.Config.BotName, config.Config.BotOAuth)
	// Join the client to the twitch channel
	client.Join(config.Config.ChannelName)

	// Waits over IRC for a message to be posted to twitch then processed here
	client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {

		isExempt := false
		for k := range user.Badges {
			if k == "moderator" || k == "broadcaster" {
				isExempt = true
			}
		}

		// If we see an ! as the first character, we can assume that it is a command. If no valid command is chosen, then we will a message sent back saying Invalid command
		if message.Text[0:1] == "!" {
			// firstSpace will get the index of the first space which should be right after the command
			firstSpace := strings.Index(message.Text, " ")
			command := message.Text[1:firstSpace]

			switch command {
			case "reg":
				regularsCommand(client, firstSpace, config.Config.ChannelName, message.Text)
				postToDiscord("Made a regulars command")
				break
			case "songs":
				songsCommand(client, firstSpace, config.Config.ChannelName, message.Text, user.Username, service, isExempt)
				break
			default:
				postToDiscord(message.Text)
			}
		}
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

// songsCommand is run when the command from twithc begins with !songs ...
func songsCommand(client *twitch.Client, firstSpace int, channelName, message, username string, service *youtube.Service, isExempt bool) {
	secondSpace := strings.Index(message[firstSpace+1:], " ")
	if secondSpace == -1 {
		if strings.Contains(message[firstSpace+1:], "list") {
			queue := currentQueue()
			if len(queue) == 0 {
				client.Say(channelName, "The queue is empty!")
			} else {
				listOfSongs := ""
				for index, song := range queue {
					listOfSongs = listOfSongs + fmt.Sprintf("%d. %s by: @%s ", index+1, song.Name, song.Who)
				}
				client.Say(channelName, listOfSongs)
			}
		} else if strings.Contains(message[firstSpace+1:], "skip") {
			if isExempt {
				commandToDiscord, messageToTwitch := skipFirstSong()
				if commandToDiscord != "" {
					postToDiscord(commandToDiscord)
				}
				client.Say(channelName, messageToTwitch)
			} else {
				// Do nothing
			}
		} else if strings.Contains(message[firstSpace+1:], "current") {
			queue := currentQueue()
			currentSong := fmt.Sprintf("Current song: %s by: @%s", queue[0].Name, queue[0].Who)
			client.Say(channelName, currentSong)
		}
	} else {
		subCommand := message[firstSpace+1 : firstSpace+1+secondSpace]
		// First two cases, we expect another argument to the command, otherwise it is invalid
		if subCommand == "add" {
			// We must now find the name and write him to the list of regulars
			name := strings.Trim(message[firstSpace+1+secondSpace:], " ")
			fmt.Printf("Adding song: %s\n", name)

			commandToDiscord, messageToTwitch := processesSongRequest(name, username, service)
			if commandToDiscord != "" {
				postToDiscord(commandToDiscord)
			}
			client.Say(channelName, messageToTwitch)
		} else if subCommand == "delete" {
			if isExempt {
				songNumber := strings.Trim(message[firstSpace+1+secondSpace:], " ")
				fmt.Printf("Removing song number: %s\n", songNumber)

				num, err := strconv.Atoi(songNumber)
				if err != nil {
					client.Say(channelName, "Invalid queue number!")
				} else {
					commandToDiscord, messageToTwitch := deleteSongRequest(num)
					if commandToDiscord != "" {
						postToDiscord(commandToDiscord)
					}
					client.Say(channelName, messageToTwitch)
				}
			} else {
				// Do nothing, this is a mod/broadcaster only command
			}
		} else {
			client.Say(channelName, "Invalid command.")
		}
	}
}

// regularsCommand is run when the command from twitch begins with !reg ...
func regularsCommand(client *twitch.Client, firstSpace int, channelName, message string) {
	secondSpace := strings.Index(message[firstSpace+1:], " ")
	// If secondSpace is -1, we will take the remaining string of the message and see if it's list.
	// If not, we will return an error message to twitch
	if secondSpace == -1 {
		if strings.Contains(message[firstSpace+1:], "list") {
			rows, _ := db.Model(&Regulars{}).Select("username").Rows()
			finalUsersString := "List of regulars:"
			for rows.Next() {
				var regular Regulars
				db.ScanRows(rows, &regular)
				finalUsersString = finalUsersString + fmt.Sprintf(" %s,", regular.Username)
			}
			finalUsersString = finalUsersString[:len(finalUsersString)-1]
			client.Say(channelName, finalUsersString)
		} else {
			client.Say(channelName, "Invalid command.")
		}
	} else {
		subCommand := message[firstSpace+1 : firstSpace+1+secondSpace]
		// First two cases, we expect another argument to the command, otherwise it is invalid
		if subCommand == "add" {
			// We must now find the name and write him to the list of regulars
			name := strings.Trim(message[firstSpace+1+secondSpace:], " ")
			fmt.Printf("Adding user: %s\n", name)

			// Add the user to the database
			reg := &Regulars{
				Username:     name,
				CurrentSongs: 0,
			}
			err := db.Create(reg).Error
			if err != nil {
				client.Say(channelName, fmt.Sprintf("%s is already a regular!", name))
			} else {
				client.Say(channelName, fmt.Sprintf("%s has been sucessfully added to the regulars list!", name))
			}

		} else if subCommand == "delete" {
			name := strings.Trim(message[firstSpace+1+secondSpace:], " ")
			fmt.Printf("Removing user: %s\n", name)

			err := db.Delete(Regulars{}, "username = ?", name).Error
			if err != nil {
				client.Say(channelName, fmt.Sprintf("Cannot delete %s from the regulars list!", name))
			} else {
				client.Say(channelName, fmt.Sprintf("%s has been sucessfully deleted from the regulars list!", name))
			}
		} else {
			client.Say(channelName, "Invalid command.")
		}
	}
}

// Checks to see if a user is a regular
func checkRegularsList(username string, isExempt bool) bool {
	if isExempt {
		return isExempt
	}
	var regular Regulars
	return !db.Where("username = ?", username).First(&regular).RecordNotFound()
}
