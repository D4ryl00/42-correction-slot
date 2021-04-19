package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Vastly inspired from https://developers.google.com/people/quickstart/go

type Me struct {
	Projects_users []ProjectUser `json:"projects_users"`
}

type ProjectUser struct {
	Project struct {
		ID int `json:"id"`
	} `json:"project"`
	Status string `json:"status"`
}

type Slot struct {
	ID    int       `json:"id"`
	Begin time.Time `json:"begin_at"`
	End   time.Time `json:"end_at"`
}

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
		//saveToken(tokFile, tok)
	}
	client := config.Client(context.Background(), tok)
	updatedToken, err := config.TokenSource(context.Background(), tok).Token()
	if err == nil {
		saveToken(tokFile, updatedToken)
	}
	return client
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

func request(client *http.Client, url string) ([]byte, error) {
	req, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	body, readErr := ioutil.ReadAll(req.Body)
	if readErr != nil {
		log.Fatal(readErr)
		return nil, err
	}

	return body, nil
}

func getProjects(client *http.Client) []int {
	req, err := request(client, "https://api.intra.42.fr/v2/me")
	if err != nil {
		return nil
	}

	var me Me
	err = json.Unmarshal(req, &me)
	if err != nil {
		log.Fatal(err)
	}

	var res []int
	for _, project := range me.Projects_users {
		if project.Status == "waiting_for_correction" {
			res = append(res, project.Project.ID)
		}
	}

	return res
}

func getProjectSlots(client *http.Client, id int) []Slot {
	var slots []Slot

	maxDay := time.Now().UTC().AddDate(0, 0, 5).Format(time.RFC3339)
	now := time.Now().UTC().Format(time.RFC3339)

	//req, err := request(client, "https://api.intra.42.fr/v2/slots")
	req, err := request(client, fmt.Sprintf("https://api.intra.42.fr/v2/projects/%d/slots?range[end_at]=%s,%s&sort=begin_at", id, now, maxDay))
	if err != nil {
		return nil
	}

	err = json.Unmarshal(req, &slots)
	if err != nil {
		log.Println("no slot found")
		return nil
	}

	return slots
}

func selectSlot(slots []Slot) *Slot {
	maxDay := time.Now().UTC().AddDate(0, 0, 5)
	minHour := 9
	minMinute := 0
	maxHour := 18
	maxMinute := 0

	for _, slot := range slots {
		if slot.End.Before(maxDay) {
			hour := slot.End.Hour()
			minute := slot.End.Minute()

			if hour >= minHour && minute >= minMinute && hour <= maxHour && minute <= maxMinute {
				return &slot
			}
		}
	}

	return nil
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

	projects := getProjects(client)
	fmt.Println("projectID:", projects)

	slots := getProjectSlots(client, projects[0])
	fmt.Println("slots:", slots)

	slot := selectSlot(slots)

	fmt.Println(slot)
}
