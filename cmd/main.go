package main

import (
	"log"
	"net"
	"net/http"

	"github.com/burhon94/NATS_sender/cmd/app"
	"github.com/burhon94/NATS_sender/pkg/configs"
	"github.com/burhon94/NATS_sender/pkg/events"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)


func main() {
	var config2 configs.Config


	err := config2.Init()
	if err != nil {
		log.Fatalf("configs err: %s", err.Error())
	}
	log.Println("configs -> Done!")
	router := mux.NewRouter()
	initStan := events.NewEvent(config2)
	stan, err := initStan.InitStan(config2)
	if err != nil {
		log.Fatalf("initStan.Error: %v", err)
	}
	server := app.NewServer(router, config2.Prefix, stan)
	server.InitRoutes()
	addr := net.JoinHostPort(config2.Host, config2.Port)
	log.Printf("server will start at address: %s", addr)
	log.Printf("endPOINT address: %s", addr+"/"+config2.Prefix)

	handler := cors.Default().Handler(server)
	loggedRouter := handlers.LoggingHandler(log.Writer(), handler)

	if err = http.ListenAndServe(addr, loggedRouter); err != nil {
		log.Println(err)
		return
	}
}
