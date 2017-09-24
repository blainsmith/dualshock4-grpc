package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/blainsmith/dualshock4-grpc/pb"
	"github.com/blainsmith/dualshock4-grpc/playstation"
	tm "github.com/buger/goterm"
	"google.golang.org/grpc"
)

// eventsService is our implementation of the protobuf interface
type eventsServer struct {
	colorchan  chan *pb.ControllerColor
	signalchan chan *pb.SignalMessage
}

// Track method receiver implements a unidirectional (Client to Server) streaming endpoint
func (es *eventsServer) Track(stream pb.Events_TrackServer) error {
	for {
		// Receive an event from the stream which is a pb.ControllerState message
		event, err := stream.Recv()

		// If it reaches the end of the stream then just return
		if err == io.EOF {
			return nil
		}

		// If we get an error from reading the stream then return the error
		if err != nil {
			return err
		}

		// Create an instance of DS4Frame
		ds4 := playstation.DS4Frame{}

		// Call UnmarshalText with the event.State to parse the byte array into a proper DS4Frame
		// Return an error if something goes wrong or the data in the byte array is invalid
		err = ds4.UnmarshalText(event.State)
		if err != nil {
			return err
		}

		// Marshal the frame into a JSON representation
		jsonds4, err := json.MarshalIndent(ds4, " ", "")
		if err != nil {
			return err
		}

		// Print the frame to the screen in JSON format
		tm.Clear()
		tm.MoveCursor(1, 1)
		tm.Print(string(jsonds4)) // Need to convert to string to print correctly
		tm.Flush()
	}
}

// Track method receiver implements a bidirectional (Server to Client) streaming endpoint
func (es *eventsServer) Color(stream pb.Events_ColorServer) error {
	// For every color we get over the eventService.colorchan we send it down to the connected client
	for color := range es.colorchan {
		stream.Send(color)
	}
	return nil
}

// Track method receiver implements a bidirectional (Client to Server) streaming endpoint
func (es *eventsServer) Signal(stream pb.Events_SignalServer) error {
	// Start a goroutine and begin sending signals from the signalchan down to the client if any
	go func() {
		for signal := range es.signalchan {
			stream.Send(signal)
		}
	}()

	// Similar as Track()
	for {
		signal, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Println(signal)
	}
}

func main() {
	// Accept a port flag from the command line, default to 1313
	port := flag.Int("port", 1313, "-port=1313")
	flag.Parse()

	// Create a TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("binding to: %d\n", *port)

	// Create our event service and initialize the channels
	es := &eventsServer{
		colorchan:  make(chan *pb.ControllerColor, 1),
		signalchan: make(chan *pb.SignalMessage, 1),
	}

	// Stream a SIGTERM signal back to the connected clients
	go func() {
		time.Sleep(20 * time.Second)
		es.signalchan <- &pb.SignalMessage{Signal: uint32(syscall.SIGTERM)}
	}()

	// Stream random colors to the controller
	go func() {
		for {
			rand.Seed(time.Now().UTC().UnixNano())

			// Continually send ControllerColor instances over the colorchan channel for .Color to send them to clients
			es.colorchan <- &pb.ControllerColor{
				Red:   uint32(rand.Intn(255)),
				Green: uint32(rand.Intn(255)),
				Blue:  uint32(rand.Intn(255)),
			}

			// Sleep every second of the loop to see the colors change
			time.Sleep(1 * time.Second)
		}
	}()

	// Create a gRPC service and register our events service with it
	grpcServer := grpc.NewServer()
	pb.RegisterEventsServer(grpcServer, es)
	grpcServer.Serve(lis)
}
