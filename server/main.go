package main

import (
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

type eventsServer struct {
	colorchan  chan *pb.ControllerColor
	signalchan chan *pb.SignalMessage
}

func (es *eventsServer) Track(stream pb.Events_TrackServer) error {
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		ds4 := playstation.DS4Frame{}
		err = ds4.UnmarshalText(event.State)
		if err != nil {
			return err
		}

		tm.MoveCursor(1, 1)
		tm.Printf("%+v", ds4)
		tm.Flush()
	}
}

func (es *eventsServer) Color(stream pb.Events_ColorServer) error {
	for color := range es.colorchan {
		stream.Send(color)
	}
	return nil
}

func (es *eventsServer) Signal(stream pb.Events_SignalServer) error {
	go func() {
		for signal := range es.signalchan {
			stream.Send(signal)
		}
	}()

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
	port := flag.Int("port", 1313, "-port=1313")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
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
			es.colorchan <- &pb.ControllerColor{
				Red:   uint32(rand.Intn(255)),
				Green: uint32(rand.Intn(255)),
				Blue:  uint32(rand.Intn(255)),
			}
			time.Sleep(1 * time.Second)
		}
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterEventsServer(grpcServer, es)
	grpcServer.Serve(lis)
}
