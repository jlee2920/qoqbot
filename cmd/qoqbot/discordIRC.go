package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/jlee2920/qoqbot.git/config"
)

// Use of discordgo as an intro to discord IRC
func startDiscordIRC() {

	discord, _ := discordgo.New("Bot " + config.Config.DiscordToken)

	discord.AddHandler(listenForMessage)
	discord.AddHandler(readyForMessages)

	_ = discord.Open()
	defer discord.Close()

}

func readyForMessages(discord *discordgo.Session, ready *discordgo.Ready) {
	_ = discord.UpdateStatus(0, "Qoqbot at your service!")

	fmt.Printf("Qoqbot has started on %+v servers", discord.State.Guilds)
}

func listenForMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// Make sure to skip everyone that is not qoqbot
	if message.Author.ID != config.Config.DiscordBotID && message.ChannelID == config.Config.DiscordServerID {
		postToDiscord(fmt.Sprintf("Message: %+v", message.Message))
	}
}

// Takes the discordURL for where to POST and the discordToken of qoqbot to echo message
func postToDiscord(message string) {
	discordURL := config.Config.DiscordURL
	discordToken := config.Config.DiscordToken
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
