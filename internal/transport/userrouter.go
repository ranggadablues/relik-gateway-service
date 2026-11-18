package transport

import (
	"net/http"

	"relik-gateway-service/controllers"
	"relik-gateway-service/internal/endpoint"
	"relik-gateway-service/middleware"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(mux *mux.Router, eps endpoint.UserEndpoints) {
	const usersId = "/users/{id}"

	publicRouter := mux.PathPrefix("/").Subrouter()

	// POST /users (create or login user)
	publicRouter.Methods("POST").Path("/users/{module}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.UserActionHandler(w, r, eps.Service)
	})

	publicRouter.Methods("GET").Path("/users/{module}/{module1}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.UserModulesHandler(w, r, eps.Service)
	})

	// POST /refresh (refresh token)
	publicRouter.Methods("POST").Path("/refresh").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.RefreshTokenHandler(w, r, eps.Service)
	})

	privateRouter := mux.PathPrefix("/").Subrouter()

	// GET /users (list users)
	privateRouter.Methods("GET").Path("/users").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.GetUsersHandler(w, r, eps.Service)
	})

	// GET /users/me (get user by JWT)
	privateRouter.Methods("GET").Path("/users/me").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.GetUserByMeHandler(w, r, eps.Service)
	})

	// GET /users/{id} (get user by id)
	privateRouter.Methods("GET").Path(usersId).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.GetUserByIdHandler(w, r, eps.Service)
	})

	// PUT /users/{id} (update user)
	privateRouter.Methods("PUT").Path(usersId).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.UpdateUserHandler(w, r, eps.Service)
	})

	// DELETE /users/{id} (delete user)
	privateRouter.Methods("DELETE").Path(usersId).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.DeleteUserHandler(w, r, eps.Service)
	})
	privateRouter.Use(middleware.AuthMiddleware)
}
