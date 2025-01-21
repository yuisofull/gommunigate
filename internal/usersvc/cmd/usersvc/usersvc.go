package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	userpb "github.com/yuisofull/gommunigate/internal/usersvc/pb"
	userendpoint "github.com/yuisofull/gommunigate/internal/usersvc/pkg/endpoint"
	"github.com/yuisofull/gommunigate/internal/usersvc/pkg/infrastructure"
	userservice "github.com/yuisofull/gommunigate/internal/usersvc/pkg/service"
	usertransport "github.com/yuisofull/gommunigate/internal/usersvc/pkg/transport"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fs := flag.NewFlagSet("usersvc", flag.ExitOnError)
	var (
		grpcAddr   = fs.String("grpc-addr", ":8081", "gRPC listen address")
		mongodbURI = fs.String("mongodb-uri", "mongodb://localhost:27017", "MongoDB URI")
		mongodbDB  = fs.String("mongodb-db", "usersvc", "MongoDB database")
		mongodbCol = fs.String("mongodb-col", "users", "MongoDB collection")
	)

	// Create a single logger, which we'll use and give to other components.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	var repo userservice.Repository
	{
		client, err := mongo.Connect(options.Client().ApplyURI(*mongodbURI))
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		logger.Log("Repository", "MongoDB", "URI", *mongodbURI, "DB", *mongodbDB, "Collection", *mongodbCol)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer client.Disconnect(ctx)
		repo = infrastructure.NewMongoRepository(client, *mongodbDB, *mongodbCol)
	}

	var (
		service    = userservice.NewService(repo)
		endpoints  = userendpoint.New(service, logger)
		grpcServer = usertransport.NewGRPCServer(endpoints, logger)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	{
		grpcListener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		defer grpcListener.Close()

		baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
		userpb.RegisterUserServer(baseServer, grpcServer)

		g.Go(func() error {
			logger.Log("transport", "gRPC", "addr", *grpcAddr)
			return baseServer.Serve(grpcListener)
		})
		defer baseServer.GracefulStop()
	}

	{
		g.Go(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			}
		})
	}
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Log("err", err)
		os.Exit(1)
	}
	logger.Log("info", "closing server gracefully")
}
