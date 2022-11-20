package routine

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/emb"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
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

		jsn, _ := emb.Cat("credentials.json")
		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(jsn, calendar.CalendarReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := GetClient(config)

		srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Printf("Unable to retrieve Calendar client: %v", err)
			return nil
		}
		//Time range for today min and max
		t := time.Now()
		min := time.Date(t.Year(), t.Month(), 21, 0, 0, 0, 0, time.Local).Format(time.RFC3339)
		max := time.Date(t.Year(), t.Month(), 21, 23, 59, 0, 0, time.Local).Format(time.RFC3339)
		events, err := srv.Events.List("primary").SingleEvents(true).TimeMin(min).TimeMax(max).OrderBy("startTime").Do()
		if err != nil {
			log.Printf("Unable to retrieve next ten of the user's events: %v", err)
			return nil
		}

		fmt.Println("Upcoming events:")
		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				s := item.Start.DateTime
				if s == "" {
					s = item.Start.Date
				}
				sTime, err := time.Parse(time.RFC3339, s)
				if err != nil {
					log.Println(err)
					continue
				}
				e := item.End.DateTime
				eTime, err := time.Parse(time.RFC3339, e)
				fmt.Printf("%v (%v-%v)\n", item.Summary, sTime.Format("3:04PM"), eTime.Format("3:04PM"))
				fmt.Printf("%v\n", item.Description)

			}
		}
		return nil
	},
}
