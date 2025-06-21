package main

import "github.com/seansa/rocket-challenge/cmd"

// @title Rocket Service API
// @version 1.0
// @description This is a service that consumes messages from rockets and exposes their via a REST API.
// @host localhost:8088
// @BasePath /
// @Schemes http
func main() {
	cmd.Run()
}
