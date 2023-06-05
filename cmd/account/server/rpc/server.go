package rpcserver

import (
	"net"

	pb "github.com/SeeJson/account/api/account"
	handler "github.com/SeeJson/account/cmd/account/handler/rpc"
	"google.golang.org/grpc"
)

type Config struct {
	Address string `mapstructure:"address"`
}

var cfg Config

func SetConfig(c Config) {
	cfg = c
}

func Run() error {
	server := &handler.Server{}
	err := start(server)
	return err
}

// Start starts server
func start(g *handler.Server) error {
	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAccountServer(grpcServer, g)
	grpcServer.Serve(lis)
	return nil
}
