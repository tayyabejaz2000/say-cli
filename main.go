package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	say "say/src"
)

func main() {
	var port = flag.Int("port", 8080, "Set Port to run app on")
	var username, _ = os.LookupEnv("USERNAME")
	var name = flag.String("name", username, "Name for the client")
	var desc = flag.String("description", "Chat App", "Description for port forwarding")
	var isLocal = flag.Bool("local", false, "Should this app run locally")
	var hidden = flag.Bool("hidden", false, "Should the name be broadcasted")
	flag.Parse()

	var appConfig = say.Config{
		Host:            flag.Args()[0] == "host",
		Name:            *name,
		BroadcastName:   *hidden,
		IsLocal:         *isLocal,
		Port:            *port,
		PortDescription: *desc,
	}
	var json, _ = json.Marshal(appConfig)
	fmt.Println(string(json))

	var app = say.CreateChatApp(&appConfig)
	var listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", "127.0.0.1", appConfig.Port))
	if err != nil {
		fmt.Printf("[Error: %s]: Error Creating Net Socket\n", err.Error())
	}
	fmt.Scanf("%s")
	fmt.Print(app, listener)
	listener.Close()
	app.Device.Close()
}
