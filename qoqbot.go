package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hpcloud/tail"
)

// JSON structure of the response we get back from the Token's API endpoint from NightBot
type nightbotTokenResp struct {
	// Structure is as follows: Name_of_variable type json_variable_in_response
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	Scope        string `json:"scope"`
}

// JSON structure of the information returned from the regulars list
type regularInfo struct {
	ID          string `json:"_id"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"2018-06-27T03:20:51.564Z"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"providerId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// Main JSON structure of the response from the Regular's API endpoint from Nightbot
type regularsResp struct {
	Total  int `json:"_total"`
	Status int `json:"status"`
	// This variable takes in the previous JSON structure as it's type
	Regulars []regularInfo `json:"regulars"`
}

func main() {

	clientID := "62e4254b14c2d05ce25bf7f384b2276e"
	redirectURI := "https://localhost/"
	authURL := "https://api.nightbot.tv/oauth2/token"

	// Get inital secrets to get the list of regulars
	// Open a reader that looks at the STDIN for input from the user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter client secret - can be found at https://beta.nightbot.tv/account/applications: ")
	// clientSecert is the var that is returned from NightBot's application that is needed. We are reading up to a newline character
	// Then trimming it off.
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)
	// Same thing we are doing for clientSecret
	fmt.Print("Enter code returned from authorizing: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)
	// Same thing we are doing for discord token
	// fmt.Print("Enter discord token: ")
	// discordToken, _ := reader.ReadString('\n')
	// discordToken = strings.TrimSpace(code)

	// Building x-www-form-urlencoded parameters. We need these parameters in this specific format because that is the only way
	// to call this API endpoint. This format is basically what you see at the end of a URL
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	// Build the request
	// We are making a new HTTP POST request to the authentication URL and adding in our parameters we just added and encoded
	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error building the http POST request: %s", err)
		return
	}
	// We create a new HTTP client to send the request and add the content type header
	client := &http.Client{}
	// Content-Type is a header that is the client telling the server what kind of data is expected to be given - required to be
	// application/x-www-form-urlencoded be NightBot
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Run the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error posting to the the tokens endpoint: %s", err)
		return
	}
	defer resp.Body.Close() //defer is a call that would happen at the end of the program no matter if it exited gracefully or not

	// Grab the response and put it into tokenResp
	// Since the response from the server is in the type io reader, we need to convert this to a different scheme in order to parse it
	// properly. To do this, we pass it into ioutil's ReadAll function to get the body of the response out into a byte array.
	// We then take this byte array, which is in JSON format, and unmarshal is into the body. Unmarshal takes a byte array that was encoded
	// from a JSON object and populated the JSON struct.
	body, err := ioutil.ReadAll(resp.Body)
	tokenResp := &nightbotTokenResp{}
	json.Unmarshal(body, tokenResp)
	resp.Body.Close()

	// Now we need to make an API post call to
	regularsURL := "https://api.nightbot.tv/1/regulars"
	req, err = http.NewRequest("GET", regularsURL, nil)
	if err != nil {
		fmt.Printf("Error building the http GET request: %s", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+tokenResp.AccessToken)

	regularsResponse, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting from regulars endpoint: %s", err)
		return
	}
	defer regularsResponse.Body.Close()
	body, err = ioutil.ReadAll(regularsResponse.Body)
	regResp := &regularsResp{}
	json.Unmarshal(body, regResp)
	regularsResponse.Body.Close()

	fmt.Printf("%q\n", regResp.Regulars)
	// Now that we have the list of regulars, we must authenticate any !play requests from twitch so that they are, in fact, a regular
	// We need an infinite loop to continually polling the log file for the !play request
	delay := time.Tick(2 * time.Second)
	for _ = range delay {
		readFile("/Users/joshualee/go/src/PhantomBot-2.4.0.3/logs/chat/01-07-2018.txt")
	}
}

func readFile(fname string) {
	t, _ := tail.TailFile(fname, tail.Config{Follow: true})
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}
