package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/jlee2920/qoqbot.git/config"

	"github.com/jinzhu/gorm"
	// import _ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// import _ "github.com/jinzhu/gorm/dialects/sqlite"
	// import _ "github.com/jinzhu/gorm/dialects/mssql"

	"github.com/caarlos0/env"

	"github.com/gempir/go-twitch-irc"
	_ "github.com/lib/pq"
)

var db *gorm.DB

func main() {
	// initialize the environment variables
	qoqbotConfig := config.GetEnv()
	// Initialize the database
	initDB(qoqbotConfig)
	defer db.Close()
	// Initiate twitch IRL client
	startTwitchIRC(qoqbotConfig)
}

// Grabs the environment variables found within the docker-compose.yml file
func setupQoqbot() conf {
	// Config is a global configuration that is used within qoqbot
	var Config conf

	if err := env.Parse(&Config); err != nil {
		panic(err)
	}
	return Config
}

type regulars struct {
	username     string
	currentSongs int
}

func initDB(qoqbot conf) {
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

	regularsBytes, _ := ioutil.ReadFile("/go/src/qoqbot/regulars.txt")
	listOfRegulars := strings.Split(string(regularsBytes), ",")

	txDB := beginTx(db)
	defer func() {
		endTx(db, err)
	}()

	for _, regular := range listOfRegulars {
		fmt.Printf("Adding to the list of regulars: %s\n", regular)
		if err := txDB.Create(&regulars{username: regular, currentSongs: 0}).Error; err != nil {
			txDB.Rollback()
			fmt.Printf("Failed to create, rolling back")
		}
	}

	fmt.Println("Finished initializing all users to database!")
}

func beginTx(conn *gorm.DB) *gorm.DB {
	conn = conn.Begin()
	if conn.Error != nil {
		fmt.Printf("Failed to start transaction\n")
	}
	return conn
}

func endTx(conn *gorm.DB, err error) {
	if err == nil {
		fmt.Printf("Committing\n")
		conn.Commit()
	} else {
		fmt.Printf("Rolling back\n")
		conn.Rollback()
	}
}

// Starts a new Twitch IRC client that listens for messages sent
func startTwitchIRC(qoqbot conf) {
	fmt.Print("Starting server!\n")
	// Instantiate a new client
	client := twitch.NewClient(qoqbot.BotName, qoqbot.BotOAuth)
	// Join the client to the twitch channel
	client.Join(qoqbot.ChannelName)

	// Waits over IRC for a message to be posted to twitch then processed here
	client.OnNewMessage(func(channel string, user twitch.User, message twitch.Message) {
		// If we see an ! as the first character, we can assume that it is a command. If no valid command is chosen, then we will a message sent back saying Invalid command
		if message.Text[0:1] == "!" {
			// firstSpace will get the index of the first space which should be right after the command
			firstSpace := strings.Index(message.Text, " ")
			command := message.Text[1:firstSpace]
			client.Say(qoqbot.ChannelName, fmt.Sprintf("Command given: %s\n", command))

			switch command {
			case "regulars":
				regularsCommand(client, firstSpace, qoqbot.ChannelName, message.Text)
				postToDiscord(qoqbot.DiscordURL, qoqbot.DiscordToken, "Made a regulars command")
				break
			case "play":
				isRegular := checkRegularsList(user.Username)
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
	// If secondSpace is -1, we will take the remaining string of the message and see if it's list. If not, we will return an error message to twitch
	if secondSpace == -1 {
		if strings.Contains(message[firstSpace+1:], "list") {
			regularsBytes, err := ioutil.ReadFile("/opt/qoqbot/regulars.txt")
			if err != nil {
				fmt.Printf("Could not find the list of regulars.\n")
				client.Say(channelName, "There are no regulars.")
			} else {
				client.Say(channelName, string(regularsBytes))
			}
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
			addingToList := "," + name

			// Get the current list
			regularsBytes, err := ioutil.ReadFile("/opt/qoqbot/regulars.txt")
			if err != nil {
				fmt.Printf("Could not find the list of regulars.\n")
				client.Say(channelName, "There are no regulars.")
			}

			// Construct the new list
			newRegularsList := string(regularsBytes) + addingToList

			// Write to the regulars file
			err = ioutil.WriteFile("/opt/qoqbot/regulars.txt", []byte(newRegularsList), 0644)
			if err != nil {
				fmt.Printf("Error adding user %s to the regulars list", name)
				client.Say(channelName, fmt.Sprintf("Could not add user %s to the regulars list", name))
			} else {
				client.Say(channelName, fmt.Sprintf("Successfully added user %s to the regulars list", name))
			}
		} else if subCommand == "remove" {
			name := strings.Trim(message[firstSpace+1+secondSpace:], " ")
			fmt.Printf("Removing user: %s\n", name)
			removeFromList := "," + name

			regularsBytes, err := ioutil.ReadFile("/opt/qoqbot/regulars.txt")
			if err != nil {
				fmt.Printf("Could not find the list of regulars.\n")
				client.Say(channelName, "There are no regulars.")
			} else {
				regulars := string(regularsBytes)
				indexOfRemove := strings.Index(regulars, removeFromList)
				if indexOfRemove == -1 {
					client.Say(channelName, fmt.Sprintf("User %s\n is not a regular.", name))
				} else {
					fmt.Printf("String: %s\nIndex of remove: %d\n", regulars, indexOfRemove)
					client.Say(channelName, regulars[:indexOfRemove]+regulars[indexOfRemove+len(removeFromList):])
				}
			}

		} else {
			client.Say(channelName, "Invalid command.")
		}
	}
}

// Reads the regulars list and searches for the username, returns true if the user exists
func checkRegularsList(username string) bool {
	regularsBytes, err := ioutil.ReadFile("/opt/qoqbot/regulars.txt")
	if err != nil {
		fmt.Printf("Could not find the list of regulars.\n")
		return false
	}
	return strings.Contains(string(regularsBytes), username)
}
