package controllers

import (
	"net/http"
	"relik-gateway-service/helper"
	"relik-gateway-service/internal/service"
	"relik-gateway-service/middleware"

	"github.com/gorilla/mux"
	"github.com/ranggadablues/gosok/auth/jwt"
	"github.com/ranggadablues/gosok/common"
	"github.com/ranggadablues/gosok/store"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	if !middleware.IsAdmin(r) {
		helper.WriteErrorHeader(w, status.Error(codes.PermissionDenied, "forbidden: admin access required"))
		return
	}

	req := &pb.ListUsersRequest{}
	resp, err := userService.GetUsers(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, resp)
}

func GetUserByIdHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	if !middleware.IsOwnerOrAdmin(r, targetID) {
		helper.WriteErrorHeader(w, status.Error(codes.PermissionDenied, "forbidden: access denied"))
		return
	}

	req := &pb.GetUserByIdRequest{
		Id: targetID,
	}

	resp, err := userService.GetUserById(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, resp)
}

func GetUserByMeHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	req := &pb.Empty{}
	ctx := jwt.InjectToGRPCContext(r.Context())
	resp, err := userService.GetUserByMe(ctx, req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}
	helper.WriteSuccessHeader(w, resp)
}

func UserActionHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	vars := mux.Vars(r)
	var response any
	var err error

	switch vars["module"] {
	case "register":
		response, err = RegisterHandler(r, userService)
	case "login":
		response, err = LoginHandler(w, r, userService)
	case "googleauth":
		GoogleAuthHandler(w, r, userService)
		return
	case "facebookauth":
		FacebookAuthHandler(w, r, userService)
		return
	case "logout":
		LogoutHandler(w, r, userService)
		return
	default:
		helper.WriteErrorHeader(w, status.Error(codes.NotFound, "unknown user action"))
		return
	}

	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, response)
}

func UserModulesHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	vars := mux.Vars(r)
	module := vars["module"]
	module1 := vars["module1"]
	var response any
	var err error

	response, err = UserVerification(r, module, module1, userService)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, response)
}

func UserVerification(r *http.Request, module, token string, userService service.IUserService) (*pb.UserVerificationResponse, error) {
	req := &pb.UserVerificationRequest{Token: token, Module: module}
	resp, err := userService.UserVerification(r.Context(), req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func RegisterHandler(r *http.Request, userService service.IUserService) (*pb.CreateBasicUserResponse, error) {
	var body pb.CreateUserRequest
	if err := common.Payload(&body, r); err != nil {
		return nil, err
	}

	resp, err := userService.CreateUser(r.Context(), &body)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) (bson.M, error) {
	req := &pb.LoginRequest{}
	if err := common.Payload(req, r); err != nil {
		return nil, err
	}
	resp, err := userService.LoginUser(r.Context(), req)
	if err != nil {
		return nil, err
	}

	setCookies(w, resp.GetToken(), resp.GetRefreshToken())
	return bson.M{"token": resp.GetToken(), "refreshToken": resp.GetRefreshToken()}, nil
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	if !middleware.IsOwnerOrAdmin(r, targetID) {
		helper.WriteErrorHeader(w, status.Error(codes.PermissionDenied, "forbidden: access denied"))
		return
	}

	var payload bson.M
	if err := common.Payload(&payload, r); err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	body := pb.UserUpdate{}
	bsonBytes, err := bson.Marshal(payload)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}
	_ = bson.Unmarshal(bsonBytes, &body)
	req := &pb.UpdateUserRequest{
		Id:   targetID,
		User: &body,
	}

	_, err = userService.UpdateUser(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, nil)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	vars := mux.Vars(r)
	targetID := vars["id"]
	if !middleware.IsOwnerOrAdmin(r, targetID) {
		helper.WriteErrorHeader(w, status.Error(codes.PermissionDenied, "forbidden: access denied"))
		return
	}

	req := &pb.DeleteUserRequest{
		Id: targetID,
	}

	resp, err := userService.DeleteUser(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	helper.WriteSuccessHeader(w, resp)
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	req := &pb.RefreshRequest{}
	if refreshCookie, err := r.Cookie("refresh_token"); err == nil && refreshCookie.Value != "" {
		req.RefreshToken = refreshCookie.Value
	} else {
		_ = common.Payload(req, r)
	}

	if req.RefreshToken == "" {
		helper.WriteErrorHeader(w, status.Error(codes.Unauthenticated, "missing refresh token"))
		return
	}

	resp, err := userService.RefreshToken(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	response := setCookies(w, resp.GetToken(), resp.GetRefreshToken())
	helper.WriteSuccessHeader(w, response)
}

func GoogleAuthHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	req := &pb.GoogleLoginRequest{}
	if err := common.Payload(req, r); err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	resp, err := userService.GoogleLogin(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	response := setCookies(w, resp.Tokens.AccessToken, resp.Tokens.RefreshToken)
	// Include user info from the request's user_info (verified by user-service)
	if req.UserInfo != nil {
		response["user"] = bson.M{
			"email":   req.UserInfo.Email,
			"name":    req.UserInfo.Name,
			"picture": req.UserInfo.Picture,
		}
	}
	helper.WriteSuccessHeader(w, response)
}

func FacebookAuthHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	req := &pb.FacebookLoginRequest{}
	if err := common.Payload(req, r); err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	resp, err := userService.FacebookLogin(r.Context(), req)
	if err != nil {
		helper.WriteErrorHeader(w, err)
		return
	}

	response := setCookies(w, resp.Tokens.AccessToken, resp.Tokens.RefreshToken)
	// Include user profile from proto response
	if resp.Profile != nil {
		response["user"] = bson.M{
			"id":    resp.Profile.Id,
			"email": resp.Profile.Email,
			"name":  resp.Profile.Name,
		}
	}
	helper.WriteSuccessHeader(w, response)
}

func setCookies(w http.ResponseWriter, accessToken string, refreshToken string) bson.M {
	store.SetCookies("access_token", accessToken, w)
	store.SetCookies("refresh_token", refreshToken, w)
	return bson.M{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, userService service.IUserService) {
	store.DeleteCookies("access_token", w)
	store.DeleteCookies("refresh_token", w)
	helper.WriteSuccessHeader(w, bson.M{"loggedOut": true})
}
