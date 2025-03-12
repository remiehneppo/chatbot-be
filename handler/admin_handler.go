package handler

import (
	"net/http"

	"github.com/tieubaoca/chatbot-be/service"
)

type AdminManageHandler interface {
	HandleCreateAdmin() http.HandlerFunc
	HandleGetAdmin() http.HandlerFunc
	HandleUpdateAdmin() http.HandlerFunc
	HandleDeleteAdmin() http.HandlerFunc
}

type adminManageHandler struct {
	adminService service.AdminService
}

func NewAdminManageHandler(adminService service.AdminService) AdminManageHandler {
	return &adminManageHandler{
		adminService: adminService,
	}
}

func (h *adminManageHandler) HandleCreateAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *adminManageHandler) HandleGetAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *adminManageHandler) HandleUpdateAdmin() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (h *adminManageHandler) HandleDeleteAdmin() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
	}
}
