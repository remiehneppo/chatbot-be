package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tieubaoca/chatbot-be/service"
	"github.com/tieubaoca/chatbot-be/types"
)

type UserManageHandler interface {
	HandleCreateUser(c *gin.Context)
	HandlerBatchCreateUser(c *gin.Context)
	HandlePaginateUser(c *gin.Context)
	HandleGetUser(c *gin.Context)
	HandleUpdateUser(c *gin.Context)
	HandleDeleteUser(c *gin.Context)
}

type userManageHandler struct {
	userService service.UserService
}

func NewUserManageHandler(userService service.UserService) UserManageHandler {
	return &userManageHandler{
		userService: userService,
	}
}

func (h *userManageHandler) HandleCreateUser(c *gin.Context) {

	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}
	user := &types.User{
		Username: req.Username,

		// Hash password
		Password:        req.Password,
		FullName:        req.FullName,
		Workspace:       req.Workspace,
		ManagementLevel: req.ManagementLevel,
		WorkspaceRole:   req.WorkspaceRole,
		CreateAt:        time.Now().Unix(),
		UpdateAt:        time.Now().Unix(),
	}
	if err := h.userService.CreateUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	res := types.DataResponse{
		Status: true,
	}
	c.JSON(http.StatusOK, res)

}

func (h *userManageHandler) HandlerBatchCreateUser(c *gin.Context) {

	var req types.BatchCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}

	users := make([]*types.User, 0)
	for _, userReq := range req.Users {
		user := &types.User{
			Username: userReq.Username,
			// Hash password
			Password:        userReq.Password,
			FullName:        userReq.FullName,
			Workspace:       userReq.Workspace,
			ManagementLevel: userReq.ManagementLevel,
			WorkspaceRole:   userReq.WorkspaceRole,
			CreateAt:        time.Now().Unix(),
			UpdateAt:        time.Now().Unix(),
		}
		users = append(users, user)
	}

	if err := h.userService.BatchCreateUser(c, users); err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}
}

func (h *userManageHandler) HandlePaginateUser(c *gin.Context) {

	var page, limit int64
	pageStr := c.Query("page")
	if pageStr == "" {
		page = 1
	} else {
		page, _ = strconv.ParseInt(pageStr, 10, 64)
	}
	limitStr := c.Query("limit")
	if limitStr == "" {
		limit = 10
	} else {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}
	users, total, err := h.userService.PaginateUser(c, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	res := types.PaginateResponse{
		Total:    total,
		Elements: users,
		Page:     page,
		Limit:    limit,
	}
	c.JSON(http.StatusOK, res)
}

func (h *userManageHandler) HandleGetUser(c *gin.Context) {

	id := c.Query("id")
	user, err := h.userService.GetUser(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	res := types.DataResponse{
		Status: true,
		Data:   user,
	}
	c.JSON(http.StatusOK, res)
}

func (h *userManageHandler) HandleUpdateUser(c *gin.Context) {
	var req types.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.DataResponse{
			Status:  false,
			Message: "Invalid request body",
		})
		return
	}

	user := &types.User{
		ID:              req.ID,
		Username:        req.Username,
		Password:        req.Password,
		FullName:        req.FullName,
		ManagementLevel: req.ManagementLevel,
		WorkspaceRole:   req.WorkspaceRole,
		Workspace:       req.Workspace,
		UpdateAt:        time.Now().Unix(),
	}

	if err := h.userService.UpdateUser(c, req.ID, user); err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}
}

func (h *userManageHandler) HandleDeleteUser(c *gin.Context) {

	id := c.Query("id")
	if err := h.userService.DeleteUser(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, types.DataResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	res := types.DataResponse{
		Status: true,
	}
	c.JSON(http.StatusOK, res)
}
