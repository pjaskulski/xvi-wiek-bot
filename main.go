package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
var homeDir string
var logFilePath string
var today string

// do testów
// var serverURL string = "http://localhost:8080"

// funkcja tworzy i zwraca instancję klienta Twittera
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

// dane z serwera api xvi-wiek.pl lokalnie (bot uruchamiany na tym samym serwerze)
func fetchData() (string, error) {

	r, err := http.Get("http://localhost:8080/api/short")
	if err != nil {
		return "", err
	}
	if r.StatusCode >= 400 {
		return "", fmt.Errorf("błąd %d. Brak danych lub niewłaściwe zapytanie", r.StatusCode)
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

	chars, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		log.Fatal(err)
	}

	fromFile := strings.Trim(string(chars), " \t")
	if fromFile == today {
		result = true
	}

	return result
}

// sprawdza i ewentualnie tworzy katalog na pliki konfiguracyjne programu
func configDir() {
	path := filepath.Join(homeDir, ".xvi-wiek-bot")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm)
		}
	}
}

// odświeża zawartość pliku log z datą ostatniego tweeta
func refreshPublished() {
	err := ioutil.WriteFile(logFilePath, []byte(today), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// ----------------------------- main ---------------------------------------
func main() {
	fmt.Println("XVI-wiek Twitter Bot")

	today = fmt.Sprintf("%04d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())

	homeDir, _ = os.UserHomeDir()
	configDir()

	logFilePath = filepath.Join(homeDir, ".xvi-wiek-bot", "xvi-wiek-bot.log")

	credentials := Credentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	}

	message, err := fetchData()
	if err != nil {
		log.Fatal(err)
	}

	if message == "" {
		log.Fatal("Brak tekstu wiadomości.")
	}

	if alreadyPublished() {
		log.Fatal("Dzisiaj już opublikowano wiadomość.")
	}

	client, err := getClient(&credentials)
	if err != nil {
		log.Println("Błąd podczas próby utworzenia klienta Twittera.")
		log.Fatal(err)
	}

	tweet, resp, err := client.Statuses.Update(message, nil)
	if err != nil {
		log.Fatal(err)
	}

	refreshPublished()

	log.Printf("%+v\n", resp)
	log.Printf("%+v\n", tweet)
}
