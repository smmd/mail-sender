package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authUrl := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Go to the following link and type the autorization code:\n", authUrl)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalln("Unable to read auth code:", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalln("Unable to retrieve token:", err)
	}

	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func sendMail(from string, to string, title string, message string, srv *gmail.Service) (bool, error) {
	// Create the message
	msgStr := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, title, message)
	msg := []byte(msgStr)
	// Get raw
	gMessage := &gmail.Message{Raw: base64.URLEncoding.EncodeToString(msg)}

	// Send the message
	_, err := srv.Users.Messages.Send("me", gMessage).Do()
	if err != nil {
		fmt.Println("Could not send mail>", err)
		return false, err
	}
	return true, nil
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalln("Unable to read client secret file:", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalln("Unable to parse client secret file config:", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalln("Unable to retrieve Gmail client:", err)
	}

	// Create message
	ret, err := sendMail("sag@test.com", "", "Gmail With Go", "It worked!", srv)
	fmt.Println("Email sent:", ret)
}
