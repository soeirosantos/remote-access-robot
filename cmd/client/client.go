package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/soeirosantos/rat/grpcapi"
	"github.com/soeirosantos/rat/pkg/util"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{grpc.WithInsecure()}

	var conn *grpc.ClientConn
	var err error
	if conn, err = grpc.Dial(util.Getenv("admin_server", "localhost:9090"), opts...); err != nil {
		log.Fatal("error starting server: ", err)
	}
	defer conn.Close()

	client := grpcapi.NewAdminClient(conn)
	req := &grpcapi.RunRequest{Command: os.Args[1]}
	res, err := client.Run(context.Background(), req)
	if err != nil {
		log.Fatal("error running command: ", err)
	}
	fmt.Println(res.Output)
}
