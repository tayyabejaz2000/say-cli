package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	say "say/src"
	"syscall"
)

func main() {
	var username, _ = os.LookupEnv("USER")
	var name = flag.String("name", username, "Name for the client (default $USER)")
	var isLocal = flag.Bool("local", false, "Run this application on local network, skips UPnP port forwarding")
	var isHost = flag.Bool("host", false, "Open TCP Port for partner to connect")
	var hidden = flag.Bool("hidden", false, "Broadcast the username to partner")
	var port = flag.Uint("port", 8080, "Set Port for TCP Socket [For running over network, this port will be forwarded by UPnP]")
	var desc = flag.String("desc", "Say", "Description for port forwarding [Ignored for running locally]")
	var help = flag.Bool("help", false, "Display Help Page")
	flag.Parse()

	if *help || len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	var appConfig = say.Config{
		Name:            *name,
		BroadcastName:   *hidden,
		IsLocal:         *isLocal,
		IsHost:          *isHost,
		Port:            uint16(*port),
		PortDescription: *desc,
	}
	var json, _ = json.Marshal(appConfig)
	fmt.Println(string(json))

	var app, err = say.CreateChatApp(&appConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-termChan
		log.Print("Ctrl-C shutting down...\n")
		app.Clean()
		os.Exit(0)
	}()

	app.Run()

	app.Clean()
}
