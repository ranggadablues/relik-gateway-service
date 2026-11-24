package transport

import (
	"io/fs"
	"net/http"
	"os"
	"relik-gateway-service/middleware"
	"strings"

	"github.com/gorilla/mux"
)

type SwaggerAssets struct {
	JSON fs.FS
	UI   fs.FS
}

func NewHTTPHandler(swagger SwaggerAssets) http.Handler {
	mux := mux.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	// Mount service-specific routes
	gateway := NewGateway()
	RegisterUserRoutes(mux, gateway.User)
	RegisterProductRoutes(mux, gateway.Product)

	// Expose Swagger UI and specs outside production, or when explicitly enabled via ENABLE_SWAGGER=true
	isProd := strings.EqualFold(os.Getenv("APP_ENV"), "prod") || strings.EqualFold(os.Getenv("APP_ENV"), "production")
	enableSwagger := strings.EqualFold(os.Getenv("ENABLE_SWAGGER"), "true")
	if !isProd || enableSwagger {
		registerSwaggerRoutes(mux, swagger)
	}
	return mux
}

func registerSwaggerRoutes(mux *mux.Router, swagger SwaggerAssets) {
	// Swagger JSON
	swaggerJSON, err := fs.Sub(swagger.JSON, "docs/swagger")
	if err != nil {
		panic("swagger JSON embed path invalid: " + err.Error())
	}
	mux.PathPrefix("/swagger/").Handler(
		http.StripPrefix("/swagger/",
			http.FileServer(http.FS(swaggerJSON)),
		),
	)

	// Swagger UI
	swaggerUI, err := fs.Sub(swagger.UI, "swagger-ui")
	if err != nil {
		panic("swagger UI embed path invalid: " + err.Error())
	}
	mux.PathPrefix("/docs/").Handler(
		http.StripPrefix("/docs/",
			http.FileServer(http.FS(swaggerUI)),
		),
	)
}
