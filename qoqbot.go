package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// The request call to get the a token from Nightbot
type nightbotTokenReq struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RedirectURI  string `json:"redirect_uri"`
	Code         string `json:"code"`
}

// Response from token API from Nightbot
type nightbotTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	Scope        string `json:"scope"`
}

type regularInfo struct {
	ID          string `json:"_id"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"2018-06-27T03:20:51.564Z"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"providerId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type regularsResp struct {
	Total    int `json:"_total"`
	Status   int `json:"status"`
	Regulars []regularInfo
}

func main() {

	clientID := "62e4254b14c2d05ce25bf7f384b2276e"
	redirectURI := "https%3A%2F%2Flocalhost%2F"
	url := "https://api.nightbot.tv/oauth2/token"

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter client secret - can be found at https://beta.nightbot.tv/account/applications: ")
	clientSecret, _ := reader.ReadString('\n')
	fmt.Print("Enter code returned from authorizing: ")
	code, _ := reader.ReadString('\n')

	getToken := &nightbotTokenReq{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		GrantType:    "authorization_code",
		RedirectURI:  redirectURI,
		Code:         code,
	}

	jsonValue, err := json.Marshal(getToken)
	if err != nil {
		fmt.Printf("Failed to marshal nightbot token request")
		return
	}

	// Build the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("Error building the http POST request: %s", err)
		return
	}
	client := &http.Client{}

	// Run the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error posting to the the tokens endpoint: %s", err)
		return
	}

	defer resp.Body.Close()

	// Use json.Decode for reading streams of JSON data
	tokenResp := &nightbotTokenResp{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		fmt.Printf("Error parsing the response from the tokens endpoint: %s", err)
		return
	}

	// Now we need to make an API post call to
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error building the http GET request: %s", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+tokenResp.AccessToken)
	client = &http.Client{}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error getting from regulars endpoint: %s", err)
		return
	}

	regularsResp := &regularsResp{}
	if err := json.NewDecoder(resp.Body).Decode(&regularsResp); err != nil {
		fmt.Printf("Error parsing the response from the regulars endpoint: %s", err)
		return
	}

	fmt.Printf("Regular count is %d", regularsResp.Total)
}
