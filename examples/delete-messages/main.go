package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tada-team/dateparse"
	"github.com/tada-team/tdclient"
	"github.com/tada-team/tdclient/examples"
	"github.com/tada-team/tdproto"
	"github.com/tada-team/tdproto/tdapi"
)

func main() {
	date := flag.String("date", "2017-12-31", "last date")

	settings := examples.NewSettings()
	settings.RequireToken()
	settings.RequireTeam()
	settings.RequireChat()
	settings.RequireDryRun()
	settings.Parse()

	dt, _ := dateparse.Parse(*date, nil)
	if dt.IsZero() {
		fmt.Println("invalid date")
		os.Exit(0)
	}

	client, err := tdclient.NewSession(settings.Server)
	if err != nil {
		panic(err)
	}

	client.SetToken(settings.Token)

	var numProcessed int

	chatUid := tdproto.JID(settings.Chat)

	filter := new(tdapi.MessageFilter)
	filter.Lang = "ru"
	filter.Limit = 200
	filter.Type = tdproto.MediatypeChange
	filter.DateTo = tdproto.IsoDatetime(dt)

	for {
		messages, err := client.GetMessages(settings.TeamUid, chatUid, filter)
		if err != nil {
			panic(err)
		}

		if len(messages) == 0 {
			break
		}

		filter.OldFrom = messages[len(messages)-1].MessageId
		for _, m := range messages {
			if !strings.HasPrefix(m.PushText, "Удалён участник:") {
				continue
			}

			numProcessed++

			if settings.DryRun {
				fmt.Println("message will be deleted (dryrun):", numProcessed, m.PushText)
				continue
			}

			if _, err := client.DeleteMessage(settings.TeamUid, chatUid, m.MessageId); err != nil {
				panic(err)
			}

			fmt.Println("message deleted:", numProcessed, m.PushText)
		}

		if len(messages) < filter.Limit {
			break
		}
	}
}
