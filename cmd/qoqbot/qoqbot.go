package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jlee2920/qoqbot.git/config"

	"github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mssql"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/gempir/go-twitch-irc"
	_ "github.com/lib/pq"
)

var db *gorm.DB

func main() {
	// initialize the environment variables
	config.InitEnv()
	// Initialize the database
	initDB(config.Config)
	defer db.Close()
	// Initiate twitch IRL client
	startTwitchIRC(config.Config)
}

// Regulars is the struct used for keeping track of who is a regular and how many songs they have done
type Regulars struct {
	ID           uint `gorm:"primary_key"`
	Username     string
	CurrentSongs int
}

func initDB(qoqbot config.Conf) {
	// Instantiate the db struct and allow db access
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		qoqbot.DBHost, qoqbot.DBPort, qoqbot.DBUser, qoqbot.DBName, qoqbot.DBPassword)

	fmt.Println(psqlInfo)
	db, err = gorm.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Printf("Failed to open gorm: %q\n", err)
	}
	fmt.Println("Successfully connected to database!")

	// Initialize all existing regulars from a text file
	fmt.Println("Reading regulars.txt to initialize all existing regulars")
	regularsBytes, _ := ioutil.ReadFile("/go/src/qoqbot.git/regulars.txt")
	listOfRegulars := strings.Split(string(regularsBytes), ",")

	for _, regular := range listOfRegulars {
		fmt.Printf("Adding to the list of regulars: %s\n", regular)
		reg := &Regulars{
			Username:     regular,
			CurrentSongs: 0,
		}
		db.Create(reg)
	}

	fmt.Println("Finished initializing all users to database!")
}

// Starts a new Twitch IRC client that listens for messages sent
func startTwitchIRC(qoqbot config.Conf) {
	fmt.Print("Starting server!\n")
	// Instantiate a new client
	client := twitch.NewClient(qoqbot.BotName, qoqbot.BotOAuth)
	// Join the client to the twitch channel
	client.Join(qoqbot.ChannelName)

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
			case "regulars":
				regularsCommand(client, firstSpace, qoqbot.ChannelName, message.Text)
				postToDiscord(qoqbot.DiscordURL, qoqbot.DiscordToken, "Made a regulars command")
				break
			case "play":
				isRegular := checkRegularsList(user.Username, isExempt)
				if isRegular {
					postToDiscord(qoqbot.DiscordURL, qoqbot.DiscordToken, message.Text)
				} else {
					client.Say(qoqbot.ChannelName, "You must be a regular request a song.\n")
				}
				break
			default:
				postToDiscord(qoqbot.DiscordURL, qoqbot.DiscordToken, message.Text)
			}
		}
	})

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

// Takes the discordURL for where to POST and the discordToken of qoqbot to echo message
func postToDiscord(discordURL, discordToken, message string) {
	httpClient := &http.Client{}
	contentBody := fmt.Sprintf(`{"content" : "%s"}`, message)
	postingJSONStruct := []byte(contentBody)

	// Creating a new POST request with the message from twitch chat
	req, err := http.NewRequest("POST", discordURL, bytes.NewBuffer(postingJSONStruct))
	if err != nil {
		fmt.Printf("Error building the http POST request: %s", err)
		return
	}

	// Creates all the authorization headers required
	req.Header.Add("Authorization", discordToken)
	req.Header.Add("Content-Type", "application/json")

	// Run the request
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error posting to the the discord's messages endpoint: %s", err)
		return
	}

	// Reads the response from discord and prints it out onto the terminal
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("POSTING TO MESSAGES: %s\n", string(body))
	resp.Body.Close()
}

// regularsCommand is run when the command from twitch begins with !regulars ...
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
