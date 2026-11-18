package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/ranggadablues/gosok/auth/jwt"
	"github.com/ranggadablues/gosok/store"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	jwtManager *jwt.Manager
	userClient pb.UserServiceClient
)

func InitAuth() error {
	manager, err := jwt.NewManagerFromEnv()
	if err != nil {
		return err
	}
	jwtManager = manager

	addr := os.Getenv("USER_SERVICE_PORT")
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	userClient = pb.NewUserServiceClient(conn)
	return nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := authenticateRequest(w, r)
		if !ok {
			return
		}

		ctx := jwt.WithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsAdmin(r *http.Request) bool {
	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok || claims == nil {
		return false
	}

	if strings.EqualFold(claims.UserInfo.Provider, "admin") {
		return true
	}

	adminEmails := os.Getenv("ADMIN_EMAILS")
	if adminEmails != "" && claims.UserInfo.Email != "" {
		for _, email := range strings.Split(adminEmails, ",") {
			if strings.EqualFold(strings.TrimSpace(email), claims.UserInfo.Email) {
				return true
			}
		}
	}

	return false
}

func IsOwnerOrAdmin(r *http.Request, targetUserID string) bool {
	if IsAdmin(r) {
		return true
	}

	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok || claims == nil {
		return false
	}

	return claims.UserInfo.UserID != "" && claims.UserInfo.UserID == targetUserID
}

func authenticateRequest(w http.ResponseWriter, r *http.Request) (*jwt.Claims, bool) {
	var rawToken string
	if accessCookie, err := r.Cookie("access_token"); err == nil && accessCookie.Value != "" {
		rawToken = accessCookie.Value
	} else if bearerToken, err := jwtManager.BearerTokenFromRequest(r); err == nil && bearerToken != "" {
		rawToken = bearerToken
	} else {
		http.Error(w, "missing access token", http.StatusUnauthorized)
		return nil, false
	}

	claims, err := jwtManager.ValidateAccessToken(rawToken)
	if err == nil {
		return claims, true
	}

	if !errors.Is(err, jwt.ErrTokenExpired) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return nil, false
	}

	return refreshAndValidate(w, r)
}

func refreshAndValidate(w http.ResponseWriter, r *http.Request) (*jwt.Claims, bool) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return nil, false
	}

	newAccess, newRefresh, err := refreshTokens(r.Context(), refreshCookie.Value)
	if err != nil {
		http.Error(w, "refresh failed", http.StatusUnauthorized)
		return nil, false
	}

	store.SetCookies("access_token", newAccess, w)
	store.SetCookies("refresh_token", newRefresh, w)

	claims, err := jwtManager.ValidateAccessToken(newAccess)
	if err != nil {
		http.Error(w, "invalid refreshed token", http.StatusUnauthorized)
		return nil, false
	}

	return claims, true
}

func refreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	resp, err := userClient.RefreshToken(ctx, &pb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return "", "", err
	}
	return resp.Token, resp.RefreshToken, nil
}
