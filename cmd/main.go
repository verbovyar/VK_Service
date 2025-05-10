package main

import (
	pb "VK_Service/ServiceApi"
	"VK_Service/config"
	"VK_Service/internal/server"
	"VK_Service/pkg/Postgresconnecting"
	"VK_Service/pkg/subpub"
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	conf, err := config.LoadConfig("././config")
	if err != nil {
		log.Fatalf("%v", err)
	}

	postgres, err := Postgresconnecting.New(conf.ConnectingString)
	if err != nil {
		logger.Fatal("pgxpool.New failed", zap.Error(err))
	}

	b := subpub.NewSubPub()
	grpcSrv := grpc.NewServer()
	srv := server.NewServer(b, postgres.Pool, logger)
	pb.RegisterPubSubServer(grpcSrv, srv)

	lis, err := net.Listen(conf.NetworkType, conf.Port)
	if err != nil {
		logger.Fatal("listen failed", zap.Error(err))
	}

	go func() {
		logger.Info("starting gRPC server")
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("Serve failed", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	logger.Info("shutting down…")

	grpcSrv.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Close(ctx); err != nil {
		logger.Warn("broker close error", zap.Error(err))
	}
}
