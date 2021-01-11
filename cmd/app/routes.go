package app

import (
	"fmt"
	"log"
)

func (s *Server) InitRoutes() {

	var routes []string
	s.router.HandleFunc(fmt.Sprintf("/%s%s", s.url, "/health"), s.handleHealth).Methods("GET")
	routes = append(routes, "/health")
	log.Printf("routes: %v", routes)

	s.router.HandleFunc(fmt.Sprintf("/%s%s", s.url, "/sender"), s.writerNATS).Methods("POST")
}
