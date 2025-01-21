package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/yuisofull/gommunigate/internal/apigateway/tokenprovider/firebase"
	userendpoint "github.com/yuisofull/gommunigate/internal/usersvc/pkg/endpoint"
	userservice "github.com/yuisofull/gommunigate/internal/usersvc/pkg/service"
	usertransport "github.com/yuisofull/gommunigate/internal/usersvc/pkg/transport"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

//var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	var (
		httpAddr                     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		userServiceInstances         = flag.String("user-service-instances", "localhost:8081", "Optional comma-separated list of URLs to user service")
		firebaseCredentialConfigFile = flag.String("firebase--credential-config-file", "/home/yui/github.com/yuisofull/gommunigate/internal/apigateway/etc/firebase-credential.json", "Firebase config file")
		retryMax                     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout                 = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	r := mux.NewRouter()

	// usersvc routes
	{
		var (
			instanceList = strings.Split(*userServiceInstances, ",")
			instancer    = sd.FixedInstancer(instanceList)
			set          userendpoint.Set
		)

		{
			factory := userSvcFactory(userendpoint.MakeGetProfileEndpoint, logger)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			set.GetProfileEndpoint = retry
		}

		{
			factory := userSvcFactory(userendpoint.MakeCreateProfileEndpoint, logger)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			set.CreateProfileEndpoint = retry
		}

		{
			factory := userSvcFactory(userendpoint.MakeUpdateProfileEndpoint, logger)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			set.UpdateProfileEndpoint = retry
		}

		{
			factory := userSvcFactory(userendpoint.MakeDeleteProfileEndpoint, logger)
			endpointer := sd.NewEndpointer(instancer, factory, logger)
			balancer := lb.NewRoundRobin(endpointer)
			retry := lb.Retry(*retryMax, *retryTimeout, balancer)
			set.DeleteProfileEndpoint = retry
		}

		userRouter := r.PathPrefix("/user").Subrouter()

		authMiddleware := AuthenticationMiddleware{TokenProvider: firebase.MustNewTokenProvider(*firebaseCredentialConfigFile)}.Middleware
		userRouter.Use(authMiddleware)

		userRouter.
			Path("/{uid}").
			Handler(httptransport.NewServer(set.GetProfileEndpoint, decodeGetProfileRequest, encodeResponse)).
			Methods(http.MethodGet)

		userRouter.
			Path("").
			Handler(httptransport.NewServer(set.CreateProfileEndpoint, decodeCreateProfileRequest, encodeResponse)).
			Methods(http.MethodPost)

		userRouter.
			Path("").
			Handler(httptransport.NewServer(set.UpdateProfileEndpoint, decodeUpdateProfileRequest, encodeResponse)).
			Methods(http.MethodPut)

		userRouter.
			Path("").
			Handler(httptransport.NewServer(set.DeleteProfileEndpoint, decodeDeleteProfileRequest, encodeResponse)).
			Methods(http.MethodDelete)
	}

	g, ctx := errgroup.WithContext(ctx)
	{
		httpListener, err := net.Listen("tcp", *httpAddr)
		if err != nil {
			logger.Log("transport", "HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		defer httpListener.Close()

		g.Go(func() error {
			logger.Log("transport", "HTTP", "addr", *httpAddr)
			return http.Serve(httpListener, r)
		})
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

	logger.Log("exit", "closing api-gateway")
}

func userSvcFactory(makeEndpoint func(userservice.Service) endpoint.Endpoint, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		service := usertransport.NewGRPCClient(conn, logger)
		endpoint := makeEndpoint(service)

		return endpoint, conn, nil
	}
}

var (
	ErrUnauthorized = errors.New("unauthorized")
)

func decodeGetProfileRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req userendpoint.GetProfileRequest
	uid := mux.Vars(r)["uid"]
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if uid != claims["user-id"] {
		req = userendpoint.GetProfileRequest{UUID: uid, Authenticated: false}
	} else {
		req = userendpoint.GetProfileRequest{UUID: uid, Authenticated: true}
	}
	return req, nil
}

func decodeCreateProfileRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req userendpoint.CreateProfileRequest
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authProvider, _ := AuthProviderFromContext(ctx)

	var request struct {
		Email       *string `json:"email"`
		PhoneNumber *string `json:"phone_number"`
		UserName    *string `json:"user_name"`
		Bio         *string `json:"bio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	uuid := claims["user-id"].(string)
	req = userendpoint.CreateProfileRequest{
		UUID:         &uuid,
		Email:        request.Email,
		PhoneNumber:  request.PhoneNumber,
		UserName:     request.UserName,
		Bio:          request.Bio,
		AuthProvider: &authProvider,
	}
	return req, nil
}

func decodeUpdateProfileRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req userendpoint.UpdateProfileRequest
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var request struct {
		Email       *string `json:"email"`
		PhoneNumber *string `json:"phone_number"`
		UserName    *string `json:"user_name"`
		Bio         *string `json:"bio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	uuid := claims["user-id"].(string)
	req = userendpoint.UpdateProfileRequest{
		UUID:        &uuid,
		Email:       request.Email,
		PhoneNumber: request.PhoneNumber,
		UserName:    request.UserName,
		Bio:         request.Bio,
	}
	return req, nil
}

func decodeDeleteProfileRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req userendpoint.DeleteProfileRequest
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	uuid := claims["user-id"].(string)
	req = userendpoint.DeleteProfileRequest{UUID: uuid}
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		encodeError(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
