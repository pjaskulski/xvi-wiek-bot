package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Credentials struct
type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

// Fact - struktura wiadomości JSON
type Fact struct {
	ContentTwitter string `json:"content"`
}

var serverURL string = "http://xvi-wiek.pl"

//var serverURL string = "http://localhost:8080"

// klient Twittera
func getClient(creds *Credentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.AccessToken, creds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// dane z serwera api xvi-wiek.pl
func fetchData() (string, error) {

	r, err := http.Get("http://localhost:8080/api/short")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	message := Fact{}
	err = json.Unmarshal(data, &message)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/dzien/%d/%d", serverURL, int(time.Now().Month()), time.Now().Day())

	return message.ContentTwitter + " " + url, nil
}

// tylko raz dziennie
func alreadyPublished() bool {
	result := false
	today := fmt.Sprintf("%04d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())

	chars, err := ioutil.ReadFile("tweets.log")
	if err != nil {
		log.Fatal(err)
	}

	fromFile := strings.Trim(string(chars), " \t")
	if fromFile == today {
		result = true
	}

	return result
}

func refreshPublished() {
	today := fmt.Sprintf("%04d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	err := ioutil.WriteFile("tweets.log", []byte(today), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("XVI-wiek Twitter Bot")

	credentials := Credentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	if alreadyPublished() {
		log.Fatal("Dzisiaj już opublikowano wiadomość.")
	}

	message, err := fetchData()
	if err != nil {
		log.Fatal(err)
	}

	if message == "" {
		log.Fatal("Brak tekstu wiadomości.")
	}

	client, err := getClient(&credentials)
	if err != nil {
		log.Println("Błąd podczas próby utworzenia klienta Twittera.")
		log.Println(err)
	}

	tweet, resp, err := client.Statuses.Update(message, nil)
	if err != nil {
		log.Println(err)
	}

	refreshPublished()

	log.Printf("%+v\n", resp)
	log.Printf("%+v\n", tweet)
}
