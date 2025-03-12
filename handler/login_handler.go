package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/tieubaoca/chatbot-be/utils"
)

type LoginHandler interface {
	HandleLogin() http.HandlerFunc
}

type loginHandler struct {
	userService service.UserService
}

func NewLoginHandler(userService service.UserService) LoginHandler {
	return &loginHandler{
		userService: userService,
	}
}

func (h *loginHandler) HandleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req types.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := h.userService.GetUserByUsername(r.Context(), req.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user.Password != req.Password {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		token, err := utils.GenerateUserToken(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := types.DataResponse{
			Status: true,
			Data: types.LoginResponse{
				AccessToken: token,
			}}
		json.NewEncoder(w).Encode(resp)
	}
}
