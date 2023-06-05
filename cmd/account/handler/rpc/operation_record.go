package rpchandler

import (
	"context"

	pb "github.com/SeeJson/account/api/account"
	radarerror "github.com/SeeJson/account/error"
	mongodao "github.com/SeeJson/account/util/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 添加操作记录rpc
func (g *Server) AddOperation(ctx context.Context, req *pb.ReqAddOperation) (*pb.RspAddOperation, error) {
	err := req.Validate()
	if err != nil {
		log.Errorf("fail to bind param: %v", err)
		return nil, &radarerror.InvalidArgs
	}

	userId := mongodao.Hex2Id(req.UserId)
	if userId == primitive.NilObjectID {
		log.Errorf("invalid UserId: %v", req.UserId)
		return nil, &radarerror.InvalidArgs
	}

	// todo
	return &pb.RspAddOperation{}, nil
}
