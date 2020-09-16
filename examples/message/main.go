package main

import (
	"log"
	"time"

	"github.com/tada-team/tdclient"
	"github.com/tada-team/tdproto"
)

func main() {
	// your team uid
	team := "uid"

	// group / direct / task chat identifier
	chat := "chat jid"

	// bot token. Type "/newbot <NAME>" command in @TadaBot direct chat
	token := "secret"

	// make http session
	client, err := tdclient.NewSession("https://web.tada.team")
	if err != nil {
		panic(err)
	}

	client.SetToken(token)

	// enable logging
	client.SetVerbose(true)

	// getting server settings, for example
	features, err := client.Features()
	if err
	log.Println("server version:", features.Build)

	// websocket connection
	ws, err := client.Ws(team, nil)
	if err != nil {
		panic(err)
	}

	// composing like human. Full events list at https://github.com/tada-team/tdproto
	ws.Send(tdproto.NewClientChatComposing(jid, true, nil))
	time.Sleep(3 * time.Second)

	// shortcut for simple messaging
	ws.SendPlainMessage(chat, "hello, world")
}
