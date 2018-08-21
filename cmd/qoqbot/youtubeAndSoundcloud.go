package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	// _ "github.com/jinzhu/gorm/dialects/mssql"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"

	_ "github.com/lib/pq"
	youtube "google.golang.org/api/youtube/v3"
	_ "gopkg.in/eapache/queue.v1"
)

// QueueElement is the struct used for keeping track of who is a regular and how many songs they have done
type QueueElement struct {
	Name    string
	EndTime time.Time
	Who     string
}

// Global queue array that mimic a queue for song requests
var queue = []QueueElement{}

func processesSongRequest(message, user string, service *youtube.Service) (string, string) {

	var songName string
	var link string
	var err error
	// If a soundcloud link is given
	if strings.Contains(message, "soundcloud.com") {
		// Contact soundcloud's API using the soundcloud link given to figure out length of time
		// Skip soundcloud for now
		return "", "Only YouTube links are supported."
	}

	// If a youtube link is given
	if strings.Contains(message, "youtube.com") {
		// This is the case where a youtube link is given
		locationOfID := strings.Index(message, "?v=") + 3
		locationOfNextParam := strings.Index(message[locationOfID:], "&")

		// Just in case there is a list param
		var id string
		if locationOfNextParam > 0 {
			id = message[locationOfID:locationOfNextParam]
		} else {
			id = message[locationOfID:]
		}

		fmt.Printf("ID %s\n", id)

		link = "!play " + message
		songName, err = addSongToQueue(service, user, id)
		if err != nil {
			return "", "Error grabbing YouTube Video."
		}
		return link, fmt.Sprintf("@%s --> Requested song: %s", user, songName)
	}

	id := getIDFromSongName(service, message)
	link = "!play https://www.youtube.com/watch?v=" + id
	songName, err = addSongToQueue(service, user, fmt.Sprintf("Song name: %s\n", message))
	if err != nil {
		return "", "Error grabbing YouTube Video."
	}
	return link, fmt.Sprintf("@%s --> Requested song: %s", user, songName)
}

func currentQueue() []QueueElement {
	// First, delete all the songs from the queue that should be removed
	deleteItemsFromQueue()
	return queue
}

func addSongToQueue(service *youtube.Service, user, id string) (string, error) {
	// First, delete all the songs from the queue that should be removed
	deleteItemsFromQueue()

	call := service.Videos.List("contentDetails,snippet")
	call = call.Id(id)

	response, err := call.Do()
	if err != nil {
		return "", err
	}
	// Time Structure: PT15M51S
	durationString := response.Items[0].ContentDetails.Duration
	titleOfSong := response.Items[0].Snippet.Title
	minuteLocation := strings.Index(durationString, "M")
	secondLocation := strings.Index(durationString, "S")

	// Get the total seconds
	minutes, _ := strconv.Atoi(durationString[2:minuteLocation])
	seconds, _ := strconv.Atoi(durationString[minuteLocation+1 : secondLocation])
	totalSeconds := time.Duration((60*minutes)+seconds) * time.Second

	// Basic element structure
	element := QueueElement{
		Name: response.Items[0].Snippet.Title,
		Who:  user,
	}

	if len(queue) == 0 {
		// If the queue is empty, we use the current server time to figure out when the song should be deleted from the queue
		element.EndTime = time.Now().Add(totalSeconds)
		queue = append(queue, element)
	} else {
		// We will use the last entry in the queue to figure out the time the next song should be deleted from the queue
		lastElement := queue[len(queue)-1]
		lastElementEndTime := lastElement.EndTime

		element.EndTime = lastElementEndTime.Add(totalSeconds)
		queue = append(queue, element)
	}

	return titleOfSong, nil
}

func deleteSongRequest(queuePlace int) (string, string) {
	// First, delete all the songs from the queue that should be removed
	deleteItemsFromQueue()

	// Make sure this queuePlace is a valid number
	if queuePlace <= len(queue)-1 && queuePlace > 0 {
		return "", fmt.Sprintf("There is no song in queue position %d.", queuePlace)
	}
	// Get the song name
	songName := queue[queuePlace-1].Name
	remove(queuePlace - 1)
	return fmt.Sprintf("!remove %d", queuePlace), fmt.Sprintf("Sucessfully removed %s.", songName)
}

func skipFirstSong() (string, string) {
	// First, delete all the songs from the queue that should be removed
	deleteItemsFromQueue()

	// Remove the first song from the queue
	if len(queue) == 0 {
		return "", "Queue is empty!"
	} else {
		songName := queue[0].Name
		remove(0)
		return "!skip", fmt.Sprintf("Skipping: %s", songName)
	}
}

func deleteItemsFromQueue() {
	currentServerTime := time.Now()

	for index, element := range queue {
		diff := currentServerTime.Sub(element.EndTime)
		if diff >= 0 {
			remove(index)
		} else {
			// Since we are appending to the queue, if we find that the diff is less than 0, all other songs in the list will also be less than 0
			return
		}
	}
}

func remove(queueNumber int) {
	queue = append(queue[:queueNumber], queue[queueNumber+1:]...)
}

func getIDFromSongName(service *youtube.Service, songName string) string {

	// Make the API call to YouTube.
	call := service.Search.List("id").
		Q(songName).
		MaxResults(1)
	response, err := call.Do()
	if err != nil {
		fmt.Printf("Error occured: %q\n", err.Error)
	}

	return response.Items[0].Id.VideoId
}
