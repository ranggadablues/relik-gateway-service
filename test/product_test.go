package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"relik-gateway-service/internal/endpoint"
	"relik-gateway-service/internal/transport"
	"relik-gateway-service/test/mocks"

	pb "github.com/ranggadablues/relik-proto-library/product-proto/pb"
)

func TestGetProductsHandler(t *testing.T) {
	mockService := &mocks.MockProductService{
		GetProductsFn: func(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
			return &pb.ListProductsResponse{
				Products: []*pb.Product{
					{Id: "1", Name: "Product A"},
					{Id: "2", Name: "Product B"},
				},
			}, nil
		},
	}

	router := mux.NewRouter()
	eps := endpoint.ProductEndpoints{Service: mockService}
	transport.RegisterProductRoutes(router, eps)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
