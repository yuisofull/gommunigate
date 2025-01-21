package main

import (
	"context"
	"flag"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/oklog/oklog/pkg/group"
	userpb "github.com/yuisofull/gommunigate/internal/usersvc/pb"
	userendpoint "github.com/yuisofull/gommunigate/internal/usersvc/pkg/endpoint"
	"github.com/yuisofull/gommunigate/internal/usersvc/pkg/infrastructure"
	userservice "github.com/yuisofull/gommunigate/internal/usersvc/pkg/service"
	usertransport "github.com/yuisofull/gommunigate/internal/usersvc/pkg/transport"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
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
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		logger.Log("repository", "MongoDB", "uri", *mongodbURI, "db", *mongodbDB, "collection", *mongodbCol)

		defer client.Disconnect(ctx)
		repo = infrastructure.NewMongoRepository(client, *mongodbDB, *mongodbCol)
	}

	var (
		service    = userservice.NewService(repo)
		endpoints  = userendpoint.New(service, logger)
		grpcServer = usertransport.NewGRPCServer(endpoints, logger)
	)

	var g group.Group
	{
		grpcListener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}

		baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
		userpb.RegisterUserServer(baseServer, grpcServer)

		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", *grpcAddr)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			baseServer.GracefulStop()
			_ = grpcListener.Close()
		})
	}

	{
		g.Add(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}, nil)
	}
	logger.Log("exit", g.Run())

}
