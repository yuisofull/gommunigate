package main

import (
	"context"
	"errors"
	"github.com/yuisofull/gommunigate/internal/apigateway/tokenprovider"
	"net/http"
)

type AuthenticationMiddleware struct {
	TokenProvider tokenprovider.TokenProvider
}

func (a *AuthenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idToken := r.Header.Get("Authorization")
		claims, err := a.TokenProvider.VerifyToken(idToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "auth_provider", a.TokenProvider.Name())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var (
	NoClaimsInContext       = errors.New("no claims in context")
	NoAuthProviderInContext = errors.New("no auth provider in context")
)

func ClaimsFromContext(ctx context.Context) (map[string]interface{}, error) {
	claims, ok := ctx.Value("claims").(map[string]interface{})
	if !ok {
		return nil, NoClaimsInContext
	}
	return claims, nil
}

func AuthProviderFromContext(ctx context.Context) (string, error) {
	authProvider, ok := ctx.Value("auth_provider").(string)
	if !ok {
		return "", NoAuthProviderInContext
	}
	return authProvider, nil
}
