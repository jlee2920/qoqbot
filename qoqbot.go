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
)

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
	Total    int           `json:"_total"`
	Status   int           `json:"status"`
	Regulars []regularInfo `json:"regulars"`
}

func main() {

	clientID := "62e4254b14c2d05ce25bf7f384b2276e"
	redirectURI := "https://localhost/"
	authURL := "https://api.nightbot.tv/oauth2/token"

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter client secret - can be found at https://beta.nightbot.tv/account/applications: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)
	fmt.Print("Enter code returned from authorizing: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	// Build the request
	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error building the http POST request: %s", err)
		return
	}
	client := &http.Client{}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Run the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error posting to the the tokens endpoint: %s", err)
		return
	}
	defer resp.Body.Close()

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
}
