package usertransport

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/yuisofull/gommunigate/internal/usersvc/pb"
	userendpoint "github.com/yuisofull/gommunigate/internal/usersvc/pkg/endpoint"
	userservice "github.com/yuisofull/gommunigate/internal/usersvc/pkg/service"
	"google.golang.org/grpc"
)

type grpcServer struct {
	createProfile grpctransport.Handler
	getProfile    grpctransport.Handler
	updateProfile grpctransport.Handler
	deleteProfile grpctransport.Handler
	pb.UnimplementedUserServer
}

// NewGRPCServer makes a set of endpoints available as a gRPC UserServer.
func NewGRPCServer(endpoints userendpoint.Set, logger log.Logger) pb.UserServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	return &grpcServer{
		createProfile: grpctransport.NewServer(
			endpoints.CreateProfileEndpoint,
			decodeGRPCCreateRequest,
			encodeGRPCCreateResponse,
			options...,
		),
		getProfile: grpctransport.NewServer(
			endpoints.GetProfileEndpoint,
			decodeGRPCRetrieveRequest,
			encodeGRPCRetrieveResponse,
			options...,
		),
		updateProfile: grpctransport.NewServer(
			endpoints.UpdateProfileEndpoint,
			decodeGRPCUpdateRequest,
			encodeGRPCUpdateResponse,
			options...,
		),
		deleteProfile: grpctransport.NewServer(
			endpoints.DeleteProfileEndpoint,
			decodeGRPCDeleteRequest,
			encodeGRPCDeleteResponse,
			options...,
		),
	}
}

func (g *grpcServer) Create(ctx context.Context, request *pb.CreateRequest) (*pb.CreateReply, error) {
	_, rep, err := g.createProfile.ServeGRPC(ctx, request)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateReply), nil
}

func (g *grpcServer) Retrieve(ctx context.Context, request *pb.RetrieveRequest) (*pb.RetrieveReply, error) {
	_, rep, err := g.getProfile.ServeGRPC(ctx, request)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.RetrieveReply), nil
}

func (g *grpcServer) Update(ctx context.Context, request *pb.UpdateRequest) (*pb.UpdateReply, error) {
	_, rep, err := g.updateProfile.ServeGRPC(ctx, request)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.UpdateReply), nil
}

func (g *grpcServer) Delete(ctx context.Context, request *pb.DeleteRequest) (*pb.DeleteReply, error) {
	_, rep, err := g.deleteProfile.ServeGRPC(ctx, request)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.DeleteReply), nil
}

func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) userservice.Service {
	var createProfileEndpoint endpoint.Endpoint
	{
		createProfileEndpoint = grpctransport.NewClient(
			conn,
			"pb.User",
			"Create",
			encodeGRPCCreateRequest,
			decodeGRPCCreateResponse,
			pb.CreateReply{},
		).Endpoint()
	}
	var getProfileEndpoint endpoint.Endpoint
	{
		getProfileEndpoint = grpctransport.NewClient(
			conn,
			"pb.User",
			"Retrieve",
			encodeGRPCRetrieveRequest,
			decodeGRPCRetrieveResponse,
			pb.RetrieveReply{},
		).Endpoint()
	}
	var updateProfileEndpoint endpoint.Endpoint
	{
		updateProfileEndpoint = grpctransport.NewClient(
			conn,
			"pb.User",
			"Update",
			encodeGRPCUpdateRequest,
			decodeGRPCUpdateResponse,
			pb.UpdateReply{},
		).Endpoint()
	}
	var deleteProfileEndpoint endpoint.Endpoint
	{
		deleteProfileEndpoint = grpctransport.NewClient(
			conn,
			"pb.User",
			"Delete",
			encodeGRPCDeleteRequest,
			decodeGRPCDeleteResponse,
			pb.DeleteReply{},
		).Endpoint()
	}
	return userendpoint.Set{
		CreateProfileEndpoint: createProfileEndpoint,
		GetProfileEndpoint:    getProfileEndpoint,
		UpdateProfileEndpoint: updateProfileEndpoint,
		DeleteProfileEndpoint: deleteProfileEndpoint,
	}
}

// decodeGRPCCreateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC create user request to a user-domain request. Primarily useful in a server.
func decodeGRPCCreateRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateRequest)
	return userendpoint.CreateProfileRequest{
		UUID:           stringPtrOrNil(req.Uuid),
		Email:          stringPtrOrNil(req.Email),
		PhoneNumber:    stringPtrOrNil(req.Phone),
		UserName:       stringPtrOrNil(req.Name),
		ProfilePicture: stringPtrOrNil(req.Profile),
		Bio:            stringPtrOrNil(req.Bio),
		AuthProvider:   stringPtrOrNil(req.AuthProvider),
	}, nil
}

// decodeGRPCRetrieveRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC retrieve user request to a user-domain request. Primarily useful in a server.
func decodeGRPCRetrieveRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.RetrieveRequest)
	return userendpoint.GetProfileRequest{UUID: stringSafeDeref(&req.Uuid)}, nil
}

// decodeGRPCUpdateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC update user request to a user-domain request. Primarily useful in a server.
func decodeGRPCUpdateRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.UpdateRequest)
	return userendpoint.UpdateProfileRequest{
		UUID:           stringPtrOrNil(req.Uuid),
		Email:          stringPtrOrNil(req.Email),
		PhoneNumber:    stringPtrOrNil(req.Phone),
		UserName:       stringPtrOrNil(req.Name),
		ProfilePicture: stringPtrOrNil(req.Profile),
		Bio:            stringPtrOrNil(req.Bio),
	}, nil
}

// decodeGRPCDeleteRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC delete user request to a user-domain request. Primarily useful in a server.
func decodeGRPCDeleteRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.DeleteRequest)
	return userendpoint.DeleteProfileRequest{UUID: stringSafeDeref(&req.Uuid)}, nil
}

// decodeGRPCCreateResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC reply to a user-domain sum response. Primarily useful in a client.
func decodeGRPCCreateResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.CreateReply)
	return userendpoint.CreateProfileResponse{Err: str2err(reply.Err)}, nil
}

// decodeGRPCRetrieveResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC reply to a user-domain sum response. Primarily useful in a client.
func decodeGRPCRetrieveResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.RetrieveReply)
	return userendpoint.GetProfileResponse{
		UUID:           stringPtrOrNil(reply.Uuid),
		Email:          stringPtrOrNil(reply.Email),
		PhoneNumber:    stringPtrOrNil(reply.Phone),
		UserName:       stringPtrOrNil(reply.Name),
		ProfilePicture: stringPtrOrNil(reply.Profile),
		Bio:            stringPtrOrNil(reply.Bio),
		Err:            str2err(reply.Err),
	}, nil
}

// decodeGRPCUpdateResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC reply to a user-domain sum response. Primarily useful in a client.
func decodeGRPCUpdateResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.UpdateReply)
	return userendpoint.UpdateProfileResponse{Err: str2err(reply.Err)}, nil
}

// decodeGRPCDeleteResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC reply to a user-domain sum response. Primarily useful in a client.
func decodeGRPCDeleteResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.DeleteReply)
	return userendpoint.DeleteProfileResponse{Err: str2err(reply.Err)}, nil
}

// encodeGRPCCreateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC create user reply. Primarily useful in a server.
func encodeGRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(userendpoint.CreateProfileResponse)
	return &pb.CreateReply{Err: err2str(resp.Err)}, nil
}

// encodeGRPCRetrieveResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC retrieve user reply. Primarily useful in a server.
func encodeGRPCRetrieveResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(userendpoint.GetProfileResponse)
	return &pb.RetrieveReply{
		Uuid:    stringSafeDeref(resp.UUID),
		Email:   stringSafeDeref(resp.Email),
		Phone:   stringSafeDeref(resp.PhoneNumber),
		Name:    stringSafeDeref(resp.UserName),
		Profile: stringSafeDeref(resp.ProfilePicture),
		Bio:     stringSafeDeref(resp.Bio),
		Err:     err2str(resp.Err),
	}, nil
}

// encodeGRPCUpdateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC update user reply. Primarily useful in a server.
func encodeGRPCUpdateResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(userendpoint.UpdateProfileResponse)
	return &pb.UpdateReply{Err: err2str(resp.Err)}, nil
}

// encodeGRPCDeleteResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC delete user reply. Primarily useful in a server.
func encodeGRPCDeleteResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(userendpoint.DeleteProfileResponse)
	return &pb.DeleteReply{Err: err2str(resp.Err)}, nil
}

// encodeGRPCCreateRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain request to a gRPC create user request. Primarily useful in a client.
func encodeGRPCCreateRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(userendpoint.CreateProfileRequest)
	return &pb.CreateRequest{
		Uuid:         stringSafeDeref(req.UUID),
		Email:        stringSafeDeref(req.Email),
		Phone:        stringSafeDeref(req.PhoneNumber),
		Name:         stringSafeDeref(req.UserName),
		Profile:      stringSafeDeref(req.ProfilePicture),
		Bio:          stringSafeDeref(req.Bio),
		AuthProvider: stringSafeDeref(req.AuthProvider),
	}, nil
}

// encodeGRPCRetrieveRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain request to a gRPC retrieve user request. Primarily useful in a client.
func encodeGRPCRetrieveRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(userendpoint.GetProfileRequest)
	return &pb.RetrieveRequest{Uuid: stringSafeDeref(&req.UUID)}, nil
}

// encodeGRPCUpdateRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain request to a gRPC update user request. Primarily useful in a client.
func encodeGRPCUpdateRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(userendpoint.UpdateProfileRequest)
	return &pb.UpdateRequest{
		Uuid:    stringSafeDeref(req.UUID),
		Email:   stringSafeDeref(req.Email),
		Phone:   stringSafeDeref(req.PhoneNumber),
		Name:    stringSafeDeref(req.UserName),
		Profile: stringSafeDeref(req.ProfilePicture),
		Bio:     stringSafeDeref(req.Bio),
	}, nil
}

// encodeGRPCDeleteRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain request to a gRPC delete user request. Primarily useful in a client.
func encodeGRPCDeleteRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(userendpoint.DeleteProfileRequest)
	return &pb.DeleteRequest{Uuid: stringSafeDeref(&req.UUID)}, nil
}

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringSafeDeref(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
