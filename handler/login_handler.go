package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/tieubaoca/chatbot-be/utils"
)

type LoginHandler interface {
	HandleLogin(c *gin.Context)
}

type loginHandler struct {
	userService service.UserService
}

func NewLoginHandler(userService service.UserService) LoginHandler {
	return &loginHandler{
		userService: userService,
	}
}

func (h *loginHandler) HandleLogin(c *gin.Context) {

	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}

	user, err := h.userService.GetUserByUsername(c, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}
	if user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Invalid password",
		})
		return
	}
	token, err := utils.GenerateUserToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}
	resp := types.DataResponse{
		Status: true,
		Data: types.LoginResponse{
			AccessToken: token,
		},
	}
	c.JSON(http.StatusOK, resp)
}
