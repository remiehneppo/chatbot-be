package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tieubaoca/chatbot-be/service"
)

type AdminManageHandler interface {
	HandleCreateAdmin(c *gin.Context)
	HandleGetAdmin(c *gin.Context)
	HandleUpdateAdmin(c *gin.Context)
	HandleDeleteAdmin(c *gin.Context)
}

type adminManageHandler struct {
	adminService service.AdminService
}

func NewAdminManageHandler(adminService service.AdminService) AdminManageHandler {
	return &adminManageHandler{
		adminService: adminService,
	}
}

func (h *adminManageHandler) HandleCreateAdmin(c *gin.Context) {

}

func (h *adminManageHandler) HandleGetAdmin(c *gin.Context) {

}

func (h *adminManageHandler) HandleUpdateAdmin(c *gin.Context) {

}

func (h *adminManageHandler) HandleDeleteAdmin(c *gin.Context) {

}
