package routine

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/emb"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

//go:embed files/*
var files embed.FS

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
	emb.FS = files
	emb.Top = "files"
}

var Cmd = &Z.Cmd{

	Name:      `routine`,
	Summary:   `a command that prints out current day routine schedule`,
	Version:   `v0.4.1`,
	Copyright: `Copyright 2022 Yongle Xu`,
	License:   `Apache-2.0`,
	Site:      `yonglexu.dev`,
	Source:    `git@github.com:rwx-yxu/routine.git`,
	Issues:    `github.com/rwx-yxu/routine/issues`,

	Commands: []*Z.Cmd{
		printCmd,
		// standard external branch imports (see rwxrob/{help,conf,vars})
		help.Cmd, conf.Cmd, vars.Cmd, emb.Cmd,
	},

	// Add custom BonzaiMark template extensions (or overwrite existing ones).

	Description: `
		{{cmd .Name}} is a tool that prints out google calendar events for the day along with any reminders. The output will be to a text file that is fed into a thermal printer. Google API key will be required to be set using routine conf edit.
			`,
}

var printCmd = &Z.Cmd{
	Name:     `today`,
	Summary:  `print today routine to standard output (default)`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		ctx := context.Background()

		jsn, _ := emb.Cat("credentials2.json")
		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(jsn, calendar.CalendarReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))

		if err != nil {
			log.Printf("Unable to retrieve Calendar client: %v", err)
			return nil
		}
		//Time min
		min := time.Now().Format(time.RFC3339)
		//max := time.Date(2022, time.November, 24, 0, 0, 0, 0, time.Local).Format(time.RFC3339)
		events, err := srv.Events.List("primary").SingleEvents(true).TimeMin(min).MaxResults(3).Do()
		if err != nil {
			log.Printf("Unable to retrieve next ten of the user's events: %v", err)
			return nil
		}
		fmt.Println("Upcoming events:")
		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				fmt.Printf("%v (%v)\n", item.Summary, date)
			}
		}
		return nil
	},
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile(config)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tok)
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
func tokenFromFile(config *oauth2.Config) (*oauth2.Token, error) {
	//Get token from bonzai conf
	c, err := Z.Conf.Data()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = yaml.Unmarshal([]byte(c), tok)
	//Refresh token if expired
	if !tok.Valid() {
		src := config.TokenSource(context.Background(), tok)
		newToken, err := src.Token()
		if err != nil {
			return tok, err
		}
		//Check that the new token is different to expired token before
		//saving to bonzai conf
		if newToken.AccessToken != tok.AccessToken {
			saveToken(newToken)
			return newToken, nil
		}
	}
	return tok, err
}

// Saves a token to bonzai config
func saveToken(token *oauth2.Token) {
	Z.Conf.OverWrite(token)
}
