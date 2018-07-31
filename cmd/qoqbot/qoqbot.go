package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jlee2920/qoqbot.git/config"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func main() {
	// initialize the environment variables
	config.InitEnv()
	// Initialize the database
	initDB()
	defer db.Close()
	// Initiate discord/twitch IRC clients
	startDiscordIRC()
	startTwitchIRC()
}

// Regulars is the struct used for keeping track of who is a regular and how many songs they have done
type Regulars struct {
	ID           uint `gorm:"primary_key"`
	Username     string
	CurrentSongs int
}

// initDB initializes the database with the use of gorm
func initDB() {
	// Instantiate the db struct and allow db access
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		config.Config.DBHost, config.Config.DBPort, config.Config.DBUser, config.Config.DBName, config.Config.DBPassword)

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
