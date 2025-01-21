package firebase

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type tokenProvider struct {
	auth *auth.Client
}

// NewTokenProvider returns a new Firebase TokenProvider,
// configured with the given config file. The config file should be a JSON file
func NewTokenProvider(configFilePath string) (*tokenProvider, error) {
	opt := option.WithCredentialsFile(configFilePath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	a, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}

	return &tokenProvider{auth: a}, nil
}

func MustNewTokenProvider(configFilePath string) *tokenProvider {
	if tp, err := NewTokenProvider(configFilePath); err != nil {
		panic(err)
	} else {
		return tp
	}
}

func (t tokenProvider) GenerateToken(data map[string]interface{}) (string, error) {
	return "", nil
}

func (t tokenProvider) VerifyToken(idToken string) (map[string]interface{}, error) {
	token, err := t.auth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return nil, err
	}
	return token.Claims, nil
}

func (t tokenProvider) Name() string {
	return "firebase"
}
