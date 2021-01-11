package events

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/burhon94/NATS_sender/pkg/configs"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

type IEvent interface {
	Publish(string, interface{}) error
	Subscribe(string, stan.MsgHandler)
}

//type Params struct {}
//
type event struct {
	conn        *nats.Conn
	encodedConn *nats.EncodedConn
	stan        stan.Conn
	config      configs.Config
}

func NewEvent(config configs.Config) *event {
	return &event{
		config: config,
	}
}

func (receiver event) InitStan(config configs.Config) (IEvent, error) {
	log.Print(config)
	var (
		URL       = config.Nats.URL // config.NatsURL
		clusterID = config.Nats.ClusterID
		clientID  = config.Nats.ClientID
	)

	// Connect to NATS
	conn, err := nats.Connect(URL,
		nats.MaxReconnects(10),
		nats.ReconnectWait(10*time.Second),
		nats.Name("NATS Streaming MOBI Subscriber"),
	)
	if err != nil {
		log.Printf("Err on Connect NATS %v", err)
	}

	// Connect to Encoded NATS
	encodedConn, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		log.Printf("Err on NewEncodedConn NATS %v", err)
	}

	// Connect to STAN
	stanConn, err := stan.Connect(clusterID, clientID,
		stan.NatsConn(conn),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v", reason)
			time.Sleep(15 * time.Second)
			conn.Close()
			_, err = receiver.InitStan(config)
		}),
	)
	if err != nil {
		log.Printf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, nats.DefaultURL)
	}
	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]", URL, clusterID, clientID)

	return &event{
		conn:        conn,        // default nats
		encodedConn: encodedConn, // encoded nats
		stan:        stanConn,    // stan
	}, err
}

//close closes nats stan connection
func (e *event) close() {
	_ = e.stan.Close()
}

//encode is the json encoding
func (e *event) encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

//decode is the json decoding // &value
func (e *event) decode(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}

//Publish
func (e *event) Publish(subject string, message interface{}) error {

	msg, err := e.encode(message)
	if err != nil {
		log.Printf("Err on publishing [%s]. Cant encode &s: %v", subject, message, err)
		return err
	}

	err = e.stan.Publish(subject, msg)
	if err != nil {
		log.Printf("Err on publishing [%s]: %v", subject, err)
		return err
	}

	log.Printf("Published [%s]: %v", subject, message)

	return err
}

//Subscribe - queue subscribe
func (e *event) Subscribe(subject string, handler stan.MsgHandler) {

	var (
		clusterID   = e.config.Nats.ClusterID
		qgroup      = "mobi"
		startSeq    uint64
		startDelta  string
		deliverAll  bool
		newOnly     bool
		deliverLast bool
		durable     = "durable"
	)

	// Process Subscriber Options.
	startOpt := stan.StartAt(pb.StartPosition_NewOnly)

	if startSeq != 0 {
		startOpt = stan.StartAtSequence(startSeq)
	} else if deliverLast {
		startOpt = stan.StartWithLastReceived()
	} else if deliverAll && !newOnly {
		startOpt = stan.DeliverAllAvailable()
	} else if startDelta != "" {
		ago, err := time.ParseDuration(startDelta)
		if err != nil {
			e.close()
			log.Printf("%v", err)
		}

		startOpt = stan.StartAtTimeDelta(ago)
	}

	_, err := e.stan.QueueSubscribe(subject, qgroup, handler, startOpt, stan.DurableName(durable))
	if err != nil {
		//e.close()
		log.Printf("Err on QueueSubscribe: %v", err)
	}

	log.Printf("Listening on [%s], clientID=[%s], qgroup=[%s] durable=[%s]\n", subject, clusterID, qgroup, durable)

	return
}

//PubRequest - request reply
func (e *event) PubRequest(subject string, req interface{}) (resp interface{}, err error) {

	if e.encodedConn != nil {
		return resp, errors.New("nats client is not connected")
	}

	b, err := e.encode(req)
	if err != nil {
		return resp, err
	}

	msg, err := e.conn.Request(subject, b, 20*time.Second)
	if err != nil {
		if e.conn.LastError() != nil {
			return resp, e.conn.LastError()
		}
		return resp, err
	}

	err = e.decode(msg.Data, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
