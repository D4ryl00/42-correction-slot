package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

// Vastly inspired from https://developers.google.com/people/quickstart/go

var clientID = flag.String("client-id", "", "the client ID")
var clientSecret = flag.String("client-secret", "", "the client secret")
var scopes = flag.String("scopes", "", "optional comma separated scopes")

func setFlags() bool {
	flag.Parse()
	missingFlag := ""

	switch {
	case *clientID == "":
		missingFlag += "client-id"
	case *clientSecret == "":
		missingFlag += "client-secret"
	default:
		return true
	}
	log.Println("setFlags: missing flags:", missingFlag)
	return false
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

func main() {
	if !setFlags() {
		os.Exit(1)
	}

	config := &oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Scopes:       strings.Split(*scopes, ","),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.intra.42.fr/oauth/authorize",
			TokenURL: "https://api.intra.42.fr/oauth/token",
		},
		RedirectURL: "https://github.com/D4ryl00/42-correction-slot",
	}

	client := getClient(config)

	fmt.Printf("etape2\n")
	req, err := client.Get("https://api.intra.42.fr/v2/cursus/42/users")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("etape3%v\n", req.Body)
	if _, err := io.Copy(os.Stdout, req.Body); err != nil {
		log.Fatal(err)
	}
}
