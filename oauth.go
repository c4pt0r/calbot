package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"google.golang.org/api/calendar/v3"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func getCalendarService(secretFile string) (*calendar.Service, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, err
	}

	srv, err := calendar.New(client)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		return nil, err
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err = saveToken(cacheFile, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser:"+
		"\n%v\nand then enter the authorization code:", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calbot.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}
