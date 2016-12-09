package main

import (
	"flag"
	"time"

	"github.com/ngaut/log"
	"google.golang.org/api/calendar/v3"
)

var (
	when     = flag.String("when", "", "When, in natural language, e.g. 下周三下午5点")
	what     = flag.String("what", "", "What")
	duration = flag.Int("dur", 1, "duration, in hour")
)

func main() {
	flag.Parse()

	if len(*when) == 0 {
		log.Fatal("missing argument `when`")
	}

	if len(*what) == 0 {
		log.Fatal("missing argument `what`")
	}

	p := newParser(&lexer{
		input: []rune(*when),
	})
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}

	tt := p.Exec()

	/*
		Use https://console.developers.google.com/start/api?id=calendar to create or select a project in the Google Developers Console and automatically turn on the API. Click Continue, then Go to credentials.
		On the Add credentials to your project page, click the Cancel button.
		At the top of the page, select the OAuth consent screen tab. Select an Email address, enter a Product name if not already set, and click the Save button.
		Select the Credentials tab, click the Create credentials button and select OAuth client ID.
		Select the application type Other, enter the name "Google Calendar API Quickstart", and click the Create button.
		Click OK to dismiss the resulting dialog.
		Click the file_download (Download JSON) button to the right of the client ID.
		Move this file to your working directory and rename it calbot.json.
	*/
	srv, err := getCalendarService("calbot.json")
	if err != nil {
		log.Fatal(err)
	}

	newEvent := calendar.Event{
		Summary: *what,
		Start:   &calendar.EventDateTime{DateTime: tt.Format(time.RFC3339)},
		End:     &calendar.EventDateTime{DateTime: tt.Add(time.Duration(*duration) * time.Hour).Format(time.RFC3339)},
	}

	_, err = srv.Events.Insert("primary", &newEvent).Do()
	if err != nil {
		log.Fatal(err)
	}
}
