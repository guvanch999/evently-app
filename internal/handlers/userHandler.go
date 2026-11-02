package handlers

import (
	"net/http"

	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// @Summary Current user
// @Description Returns the profile of the authenticated user.
// @Tags Users
// @Produce json
// @Security Bearer
// @Success 200 {object} util.DataWrapper[dto.UserResponseDto]
// @Failure 401 {object} exceptions.ResponseError
// @Router /me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userId, err := util.GetUserId(c)
	if err != nil {
		return err
	}
	user, err := h.service.GetProfile(userId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, util.DataWrapper[*dto.UserResponseDto]{Data: user})
}

// @Summary Update current user
// @Description Partially updates the authenticated user's profile. Send an empty birth_date to clear it.
// @Tags Users
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body dto.UpdateProfileDto true "Fields to update"
// @Success 200 {object} util.DataWrapper[dto.UserResponseDto]
// @Failure 400 {object} exceptions.ResponseError
// @Failure 401 {object} exceptions.ResponseError
// @Failure 422 {object} exceptions.ResponseError
// @Router /me [patch]
func (h *UserHandler) UpdateMe(c echo.Context) error {
	userId, err := util.GetUserId(c)
	if err != nil {
		return err
	}
	var body dto.UpdateProfileDto
	if err := c.Bind(&body); err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}
	if err := c.Validate(&body); err != nil {
		return err
	}
	user, err := h.service.UpdateProfile(userId, body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, util.DataWrapper[*dto.UserResponseDto]{Data: user})
}
