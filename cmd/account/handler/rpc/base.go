package rpchandler

import (
	pb "github.com/SeeJson/account/api/account"
)

type Server struct {
	pb.UnimplementedAccountServer
}
