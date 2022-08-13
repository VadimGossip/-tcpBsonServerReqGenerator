package main

import "github.com/VadimGossip/tcpBsonServerReqGenerator/internal/app"

var configDir = "config"

func main() {
	app.Run(configDir)
}
