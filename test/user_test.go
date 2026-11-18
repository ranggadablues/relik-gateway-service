package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"relik-gateway-service/internal/endpoint"
	"relik-gateway-service/internal/transport"
	"relik-gateway-service/test/mocks"

	"github.com/gorilla/mux"
	pb "github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

func TestGetUserSuccess(t *testing.T) {
	const testEmail = "test@mail.com"
	mockUserService := &mocks.MockUserService{
		GetUsersFn: func(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
			return &pb.ListUsersResponse{Users: []*pb.User{{Id: "123", Email: testEmail}}}, nil
		},
		GetUserByIdFn: func(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
			return &pb.GetUserByIdResponse{User: &pb.User{Id: "123", Email: testEmail}}, nil
		},
		CreateUserFn: func(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateBasicUserResponse, error) {
			return &pb.CreateBasicUserResponse{Link: "fake-link"}, nil
		},
		UpdateUserFn: func(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
			return &pb.UpdateUserResponse{Success: true}, nil
		},
		DeleteUserFn: func(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
			return &pb.DeleteUserResponse{Success: true}, nil
		},
		LoginUserFn: func(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
			return &pb.LoginResponse{
				Token:        "fake-token-123",
				RefreshToken: "fake-refresh-456",
			}, nil
		},
	}

	router := mux.NewRouter()
	eps := endpoint.MakeUserEndpoints(mockUserService)
	transport.RegisterUserRoutes(router, eps)

	body := `{"email":"test@mail.com","password":"123"}`
	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	resp, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T", response["data"])
	}

	if resp["token"] != "fake-token-123" {
		t.Fatalf("unexpected token: %v", resp["token"])
	}
	if resp["refreshToken"] != "fake-refresh-456" {
		t.Fatalf("unexpected refresh token: %v", resp["refreshToken"])
	}
}
