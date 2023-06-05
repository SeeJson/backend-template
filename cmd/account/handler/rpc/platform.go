package rpchandler

import (
	"context"

	pb "github.com/SeeJson/account/api/account"
	radarerror "github.com/SeeJson/account/error"
	log "github.com/sirupsen/logrus"
)

// 添加操作记录rpc
func (g *Server) GetsPlatform(ctx context.Context, req *pb.ReqGetsPlatform) (*pb.RspGetsPlatform, error) {
	err := req.Validate()
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		return nil, &radarerror.InvalidArgs
	}
	// todo
	return &pb.RspGetsPlatform{}, nil
}
