package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const queryStr = "from:Shinken-monitoring  newer_than:1d"

// can use this in the future to
// build more granular query
type message struct {
	time     int64
	gmail_ID string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
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

// from:Shinken-monitoring  newer_than:1d
func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	//user := "me"
	//msg := []message{}
	//pageToken := ""
	var sendMess gmail.Message

	// basic loop to check if there are any status alerts
	// in 24 hour time, and if no send email warning

	for {

		req := srv.Users.Messages.List("me").Q(queryStr)

		r, err := req.Do()

		if err != nil {
			log.Fatalf("unable to retrieve messages: %s", err)
		}

		numMess := len(r.Messages)

		if numMess == 0 {
			// send a warning message
			messageStr := []byte(
				"From: reburns@protonmail.com\r\n" +
					"To: bossman@checker.com\r\n" +
					"Subject: Possible Monitoring Failure\r\n\r\n" +
					"Hey Bob!\nThere may be an issue with Shinken. I have not recieved any alerts for 24 hours.\n\nPeace.")

			sendMess.Raw = base64.URLEncoding.EncodeToString(messageStr)

			// Send the message
			_, err = srv.Users.Messages.Send("me", &sendMess).Do()
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Warning Message sent")
			}

		} else {
			log.Printf("%d messages received in the last 24 hours\n", numMess)
		}

		log.Println("Sleep for 1 hour")
		time.Sleep(time.Hour * 1)

	}

}
