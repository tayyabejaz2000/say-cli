package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	say "say/src"
)

func main() {
	var username, _ = os.LookupEnv("Username for Application")
	var name = flag.String("name", username, "Name for the client")
	var isLocal = flag.Bool("local", false, "Run this application on local network, skips UPnP port forwarding")
	var hidden = flag.Bool("hidden", false, "Broadcast the username to partner")
	var port = flag.Uint("port", 8080, "Set Port for TCP Socket [For running over network, this port will be forwarded by UPnP]")
	var desc = flag.String("desc", "Say App", "Description for port forwarding [Ignored for running locally]")
	flag.Parse()

	var appConfig = say.Config{
		Name:            *name,
		BroadcastName:   *hidden,
		IsLocal:         *isLocal,
		Port:            uint16(*port),
		PortDescription: *desc,
	}
	var json, _ = json.Marshal(appConfig)
	fmt.Println(string(json))

	var app, err = say.CreateChatApp(&appConfig)
	if err != nil {
		os.Exit(-1)
	}

	app.Run()

	app.Clean()
}
