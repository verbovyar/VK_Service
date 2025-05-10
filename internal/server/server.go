package server

import (
	pb "VK_Service/ServiceApi"
	"VK_Service/internal/repositories/interfaces"
	"VK_Service/pkg/subpub"
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedPubSubServer
	broker subpub.SubPub
	log    *zap.Logger
	db     interfaces.Iface
}

func NewServer(b subpub.SubPub, db interfaces.Iface, log *zap.Logger) *Server {
	return &Server{broker: b, db: db, log: log}
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	sub, err := s.broker.Subscribe(req.Key, func(msg interface{}) {
		ev := &pb.Event{Data: msg.(string)}
		if err := stream.Send(ev); err != nil {
			s.log.Error("send to client failed", zap.Error(err))
		}
	})
	if err != nil {
		return status.Errorf(codes.Unavailable, "can't subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	err, rows := s.db.Subscribe(req, stream)

	if err != nil {
		s.log.Error("history query failed", zap.Error(err))
	} else {
		var data string
		for rows.Next() {
			rows.Scan(&data)
			stream.Send(&pb.Event{Data: data})
		}
	}

	<-stream.Context().Done()
	return stream.Context().Err()
}

func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	err := s.db.Publish(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "db insert failed: %v", err)
	}

	if err = s.broker.Publish(req.Key, req.Data); err != nil {
		return nil, status.Errorf(codes.Internal, "publish failed: %v", err)
	}
	return &emptypb.Empty{}, nil
}
