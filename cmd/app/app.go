package app

import (
	"encoding/json"
	nats "github.com/burhon94/NATS_sender/pkg/events"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const MAX_UPLOAD_FILE_SIZE = 10 * 1024 * 1024 // 10mb;
type Server struct {
	router *mux.Router
	url    string
	iEvent nats.IEvent
}

func NewServer(router *mux.Router, url string, stan nats.IEvent) *Server {
	return &Server{router: router, url: url, iEvent: stan}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

func (s *Server) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("It sender, work!"))
}

type SendSMSData struct {
	User    userStruct
	Message string
}

type userStruct struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (s *Server) writerNATS(w http.ResponseWriter, r *http.Request) {

	var (
		//response structs.Response
		user userStruct
	)

	bodyReader, err2 := ioutil.ReadAll(r.Body)
	if err2 != nil {
		panic(err2)
	}

	err2 = json.Unmarshal(bodyReader, &user)
	if err2 != nil {
		panic(err2)
	}

	// send sms
	var smsData SendSMSData
	smsData.User = user
	smsData.Message = "Hello from NATS"

	err := s.iEvent.Publish("test1", smsData)
	if err != nil {
		log.Printf("Err Publish [send-profile-sms]: %+v\n", err)
	}
	log.Print("After sending sms event")
}

/*

	//// publish foo event
	//err = h.event.Publish("foo", []byte("start"))
	//if err != nil {
	//	h.logger.Error("Err Publish [foo]: %+v\n", err)
	//}
	//
	//me := &person{Name: "Derek", Age: 0, Address: utils.GenRandNum() + " San Francisco, CA"}
	//
	//h.logger.InfoWithSpan(span, "Before sending loop hello event")
	//// firing 50 events
	//for i := 0; i < 50; i++ {
	//	me.Age = i
	//
	//	// fire hello event
	//	err := h.event.Publish("hello", me)
	//	if err != nil {
	//		h.logger.Error("Err Publish: %+v\n", err)
	//	}
	//}
	//h.logger.InfoWithSpan(span, "After sending loop hello event")

	//// publish foo event
	//_ = h.event.Publish("foo", []byte("stop"))
	//
	//response = responses.Success
 */