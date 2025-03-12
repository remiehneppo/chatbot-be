package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

type UserManageHandler interface {
	HandleCreateUser() http.HandlerFunc
	HandlerBatchCreateUser() http.HandlerFunc
	HandlePaginateUser() http.HandlerFunc
	HandleGetUser() http.HandlerFunc
	HandleUpdateUser() http.HandlerFunc
	HandleDeleteUser() http.HandlerFunc
}

type userManageHandler struct {
	userService service.UserService
}

func NewUserManageHandler(userService service.UserService) UserManageHandler {
	return &userManageHandler{
		userService: userService,
	}
}

func (h *userManageHandler) HandleCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req types.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		user := &types.User{
			Username: req.Username,

			// Hash password
			Password:        req.Password,
			FullName:        req.FullName,
			Role:            req.Role,
			Workspace:       req.Workspace,
			ManagementLevel: req.ManagementLevel,
			WorkspaceRole:   req.WorkspaceRole,
			CreateAt:        time.Now().Unix(),
			UpdateAt:        time.Now().Unix(),
		}
		if err := h.userService.CreateUser(r.Context(), user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res := types.DataResponse{
			Status: true,
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (h *userManageHandler) HandlerBatchCreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req types.BatchCreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		users := make([]*types.User, 0)
		for _, userReq := range req.Users {
			user := &types.User{
				Username: userReq.Username,

				// Hash password
				Password:        userReq.Password,
				FullName:        userReq.FullName,
				Role:            userReq.Role,
				Workspace:       userReq.Workspace,
				ManagementLevel: userReq.ManagementLevel,
				WorkspaceRole:   userReq.WorkspaceRole,
				CreateAt:        time.Now().Unix(),
				UpdateAt:        time.Now().Unix(),
			}
			users = append(users, user)
		}

		if err := h.userService.BatchCreateUser(r.Context(), users); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *userManageHandler) HandlePaginateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var page, limit int64
		pageStr := r.URL.Query().Get("page")
		if pageStr == "" {
			page = 1
		} else {
			page, _ = strconv.ParseInt(pageStr, 10, 64)
		}
		limitStr := r.URL.Query().Get("limit")
		if limitStr == "" {
			limit = 10
		} else {
			limit, _ = strconv.ParseInt(limitStr, 10, 64)
		}
		users, total, err := h.userService.PaginateUser(r.Context(), page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res := types.PaginateResponse{
			Total:    total,
			Elements: users,
			Page:     page,
			Limit:    limit,
		}
		json.NewEncoder(w).Encode(res)
	}
}

func (h *userManageHandler) HandleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("id")
		user, err := h.userService.GetUser(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func (h *userManageHandler) HandleUpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req types.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user := &types.User{
			ID:              req.ID,
			Username:        req.Username,
			Password:        req.Password,
			FullName:        req.FullName,
			ManagementLevel: req.ManagementLevel,
			Role:            req.Role,
			WorkspaceRole:   req.WorkspaceRole,
			Workspace:       req.Workspace,
			UpdateAt:        time.Now().Unix(),
		}

		if err := h.userService.UpdateUser(r.Context(), req.ID, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func (h *userManageHandler) HandleDeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Query().Get("id")
		if err := h.userService.DeleteUser(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res := types.DataResponse{
			Status: true,
		}
		json.NewEncoder(w).Encode(res)
	}
}
