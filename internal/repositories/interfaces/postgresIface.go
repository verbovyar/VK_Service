package interfaces

import (
	pb "VK_Service/ServiceApi"
	"context"
	"github.com/jackc/pgx/v5"
)

type Iface interface {
	Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) (error, pgx.Rows)
	Publish(ctx context.Context, req *pb.PublishRequest) error
}
