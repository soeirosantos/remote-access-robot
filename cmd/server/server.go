package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/soeirosantos/rat/grpcapi"
	"github.com/soeirosantos/rat/pkg/util"
	"google.golang.org/grpc"
)

type robotServer struct {
	command, output chan *string
}

type adminServer struct {
	command, output chan *string
}

func NewRobotServer(command, output chan *string) *robotServer {
	return &robotServer{command: command, output: output}
}

func NewAdminServer(command, output chan *string) *adminServer {
	return &adminServer{command: command, output: output}
}

func (s *robotServer) Fetch(ctx context.Context, empty *grpcapi.FetchRequest) (*grpcapi.FetchResponse, error) {
	select {
	case cmd, ok := <-s.command:
		if ok {
			return &grpcapi.FetchResponse{Command: *cmd}, nil
		}
		return nil, errors.New("channel closed")
	default:
		// noop
		return &grpcapi.FetchResponse{}, nil
	}
}

func (s *robotServer) Send(ctx context.Context, result *grpcapi.SendRequest) (*grpcapi.SendResponse, error) {
	s.output <- &result.Output
	return &grpcapi.SendResponse{}, nil
}

func (s *adminServer) Run(ctx context.Context, cmd *grpcapi.RunRequest) (*grpcapi.RunResponse, error) {
	go func() {
		s.command <- &cmd.Command
	}()
	var res *string
	res = <-s.output
	return &grpcapi.RunResponse{Output: *res}, nil
}

func main() {

	command, output := make(chan *string), make(chan *string)
	robot := NewRobotServer(command, output)
	admin := NewAdminServer(command, output)

	var err error

	var robotListener net.Listener
	if robotListener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", util.Getenv("robot_port", "9091"))); err != nil {
		log.Fatal(err)
	}

	var adminListener net.Listener
	if adminListener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", util.Getenv("admin_port", "9090"))); err != nil {
		log.Fatal(err)
	}

	opts := []grpc.ServerOption{}
	grpcAdminServer, grpcRobotServer := grpc.NewServer(opts...), grpc.NewServer(opts...)

	grpcapi.RegisterRobotServer(grpcRobotServer, robot)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := grpcRobotServer.Serve(robotListener); err != nil {
			log.Fatal("error serving gRPC robot ")
		}
	}()

	go func() {
		if err := grpcAdminServer.Serve(adminListener); err != nil {
			log.Fatal("error serving gRPC admin ")
		}
	}()

	log.Println("servers started")

	<-done

	log.Println("servers stopped")
}
