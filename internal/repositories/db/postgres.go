package db

import (
	pb "VK_Service/ServiceApi"
	"VK_Service/internal/repositories/interfaces"
	"context"
	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	Pool interfaces.PgxPoolIface
}

func New(pool interfaces.PgxPoolIface) *PostgresRepository {
	return &PostgresRepository{
		Pool: pool,
	}
}

func (r *PostgresRepository) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) (error, pgx.Rows) {
	rows, err := r.Pool.Query(stream.Context(),
		`SELECT data FROM messages WHERE subject=$1 ORDER BY created_at`,
		req.Key,
	)
	if err != nil {
		return err, nil
	}

	return nil, rows
}

func (r *PostgresRepository) Publish(ctx context.Context, req *pb.PublishRequest) error {
	_, err := r.Pool.Exec(ctx,
		`INSERT INTO messages (subject, data, created_at) VALUES ($1, $2, NOW())`,
		req.Key, req.Data,
	)

	return err
}
