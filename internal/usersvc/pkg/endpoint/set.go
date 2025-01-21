package userendpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/yuisofull/gommunigate/internal/usersvc/pkg/model"
	userservice "github.com/yuisofull/gommunigate/internal/usersvc/pkg/service"
)

type Set struct {
	CreateProfileEndpoint endpoint.Endpoint
	GetProfileEndpoint    endpoint.Endpoint
	UpdateProfileEndpoint endpoint.Endpoint
	DeleteProfileEndpoint endpoint.Endpoint
}

func New(s userservice.Service, logger log.Logger) Set {
	return Set{
		CreateProfileEndpoint: MakeCreateProfileEndpoint(s),
		GetProfileEndpoint:    MakeGetProfileEndpoint(s),
		UpdateProfileEndpoint: MakeUpdateProfileEndpoint(s),
		DeleteProfileEndpoint: MakeDeleteProfileEndpoint(s),
	}
}

// CreateProfile implements Service. Primarily useful in a client.
func (s Set) CreateProfile(ctx context.Context, u model.User) error {
	request := CreateProfileRequest{
		UUID:           u.UUID,
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		UserName:       u.UserName,
		ProfilePicture: u.ProfilePicture,
		Bio:            u.Bio,
		AuthProvider:   u.AuthProvider,
	}
	response, err := s.CreateProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(CreateProfileResponse)
	return resp.Err
}

func (s Set) GetProfile(ctx context.Context, uid string, authenticated bool) (model.User, error) {
	request := GetProfileRequest{UUID: uid}
	response, err := s.GetProfileEndpoint(ctx, request)
	if err != nil {
		return model.User{}, err
	}
	resp := response.(GetProfileResponse)
	return model.User{
		UUID:           resp.UUID,
		Email:          resp.Email,
		PhoneNumber:    resp.PhoneNumber,
		UserName:       resp.UserName,
		ProfilePicture: resp.ProfilePicture,
		Bio:            resp.Bio,
	}, resp.Err
}

func (s Set) UpdateProfile(ctx context.Context, u model.User) error {
	request := UpdateProfileRequest{
		UUID:           u.UUID,
		Email:          u.Email,
		PhoneNumber:    u.PhoneNumber,
		UserName:       u.UserName,
		ProfilePicture: u.ProfilePicture,
		Bio:            u.Bio,
	}
	response, err := s.UpdateProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(UpdateProfileResponse)
	return resp.Err
}

func (s Set) DeleteProfile(ctx context.Context, uid string) error {
	request := DeleteProfileRequest{UUID: uid}
	response, err := s.DeleteProfileEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(DeleteProfileResponse)
	return resp.Err
}

func MakeCreateProfileEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateProfileRequest)
		err = s.CreateProfile(ctx, model.User{
			UUID:           req.UUID,
			Email:          req.Email,
			PhoneNumber:    req.PhoneNumber,
			UserName:       req.UserName,
			ProfilePicture: req.ProfilePicture,
			Bio:            req.Bio,
			AuthProvider:   req.AuthProvider,
		})
		return CreateProfileResponse{Err: err}, nil
	}
}

func MakeGetProfileEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetProfileRequest)
		u, err := s.GetProfile(ctx, req.UUID, false)
		return GetProfileResponse{
			UUID:           u.UUID,
			Email:          u.Email,
			PhoneNumber:    u.PhoneNumber,
			UserName:       u.UserName,
			ProfilePicture: u.ProfilePicture,
			Bio:            u.Bio,
			Err:            err,
		}, nil
	}
}

func MakeUpdateProfileEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateProfileRequest)
		err = s.UpdateProfile(ctx, model.User{
			UUID:           req.UUID,
			Email:          req.Email,
			PhoneNumber:    req.PhoneNumber,
			UserName:       req.UserName,
			ProfilePicture: req.ProfilePicture,
			Bio:            req.Bio,
		})
		return UpdateProfileResponse{Err: err}, nil
	}
}

func MakeDeleteProfileEndpoint(s userservice.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteProfileRequest)
		err = s.DeleteProfile(ctx, req.UUID)
		return DeleteProfileResponse{Err: err}, nil
	}

}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = CreateProfileResponse{}
	_ endpoint.Failer = GetProfileResponse{}
	_ endpoint.Failer = UpdateProfileResponse{}
	_ endpoint.Failer = DeleteProfileResponse{}
)

// CreateProfileRequest collects the request parameters for the CreateProfile method.
type CreateProfileRequest struct {
	UUID           *string `json:"uid"`
	Email          *string `json:"email,omitempty"`
	PhoneNumber    *string `json:"phoneNumber,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	AuthProvider   *string `json:"authProvider,omitempty"`
}

// CreateProfileResponse collects the response values for the CreateProfile method.
type CreateProfileResponse struct {
	Err error `json:"-"`
}

// Failed implements endpoint.Failer.
func (r CreateProfileResponse) Failed() error { return r.Err }

// GetProfileRequest collects the request parameters for the GetProfile method.
type GetProfileRequest struct {
	UUID          string
	Authenticated bool
}

// GetProfileResponse collects the response values for the GetProfile method.
type GetProfileResponse struct {
	UUID           *string `json:"uid"`
	Email          *string `json:"email,omitempty"`
	PhoneNumber    *string `json:"phoneNumber,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	Err            error   `json:"-"`
}

// Failed implements endpoint.Failer.
func (r GetProfileResponse) Failed() error { return r.Err }

// UpdateProfileRequest collects the request parameters for the UpdateProfile method.
type UpdateProfileRequest struct {
	UUID           *string `json:"uid"`
	Email          *string `json:"email,omitempty"`
	PhoneNumber    *string `json:"phoneNumber,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Bio            *string `json:"bio,omitempty"`
}

// UpdateProfileResponse collects the response values for the UpdateProfile method.
type UpdateProfileResponse struct {
	Err error `json:"-"`
}

// Failed implements endpoint.Failer.
func (r UpdateProfileResponse) Failed() error { return r.Err }

// DeleteProfileRequest collects the request parameters for the DeleteProfile method.
type DeleteProfileRequest struct {
	UUID string `json:"uid"`
}

// DeleteProfileResponse collects the response values for the DeleteProfile method.
type DeleteProfileResponse struct {
	Err error `json:"-"`
}

// Failed implements endpoint.Failer.
func (r DeleteProfileResponse) Failed() error { return r.Err }
