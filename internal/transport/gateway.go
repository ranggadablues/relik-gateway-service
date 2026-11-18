package transport

import (
	"os"

	"relik-gateway-service/internal/service"

	"relik-gateway-service/internal/endpoint"
)

type Gateway struct {
	User    endpoint.UserEndpoints
	Product endpoint.ProductEndpoints
	// add more services here
}

func NewGateway() Gateway {
	userServicePort := os.Getenv("USER_SERVICE_PORT")
	productServicePort := os.Getenv("PRODUCT_SERVICE_PORT")
	return Gateway{
		User:    endpoint.MakeUserEndpoints(service.NewUserService(userServicePort)),
		Product: endpoint.MakeProductEndpoints(service.NewProductService(productServicePort)),
	}
}
