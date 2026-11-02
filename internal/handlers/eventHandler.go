package handlers

import (
	"net/http"

	"github.com/guvanch999/evently-app/internal/dto"
	"github.com/guvanch999/evently-app/internal/services"
	"github.com/guvanch999/evently-app/utils/exceptions"
	"github.com/guvanch999/evently-app/utils/util"
	"github.com/labstack/echo/v4"
)

type EventHandler struct {
	service *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// @Summary List events
// @Description Returns events ordered by start time, paginated.
// @Tags Events
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 20, max 100)"
// @Success 200 {object} util.DataWrapper[dto.EventListDto]
// @Failure 400 {object} exceptions.ResponseError
// @Router /events [get]
func (h *EventHandler) ListEvents(c echo.Context) error {
	var query dto.PaginationQuery
	if err := c.Bind(&query); err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}
	if err := c.Validate(&query); err != nil {
		return err
	}
	events, err := h.service.ListEvents(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, util.DataWrapper[*dto.EventListDto]{Data: events})
}

// @Summary Get event
// @Tags Events
// @Produce json
// @Param id path string true "Event id (UUID)"
// @Success 200 {object} util.DataWrapper[dto.EventResponseDto]
// @Failure 400 {object} exceptions.ResponseError
// @Failure 404 {object} exceptions.ResponseError
// @Router /events/{id} [get]
func (h *EventHandler) GetEvent(c echo.Context) error {
	id, err := util.GetParamsUUID(c, "id")
	if err != nil {
		return err
	}
	event, err := h.service.GetEvent(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, util.DataWrapper[*dto.EventResponseDto]{Data: event})
}

// @Summary Create event
// @Description Creates an event owned by the authenticated user.
// @Tags Events
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body dto.CreateEventDto true "Event"
// @Success 201 {object} util.DataWrapper[dto.EventResponseDto]
// @Failure 400 {object} exceptions.ResponseError
// @Failure 401 {object} exceptions.ResponseError
// @Failure 422 {object} exceptions.ResponseError
// @Router /events [post]
func (h *EventHandler) CreateEvent(c echo.Context) error {
	userId, err := util.GetUserId(c)
	if err != nil {
		return err
	}
	var body dto.CreateEventDto
	if err := c.Bind(&body); err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}
	if err := c.Validate(&body); err != nil {
		return err
	}
	event, err := h.service.CreateEvent(userId, body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, util.DataWrapper[*dto.EventResponseDto]{Data: event})
}

// @Summary Update event
// @Description Partially updates an event. Only the owner may update it.
// @Tags Events
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Event id (UUID)"
// @Param body body dto.UpdateEventDto true "Fields to update"
// @Success 200 {object} util.DataWrapper[dto.EventResponseDto]
// @Failure 400 {object} exceptions.ResponseError
// @Failure 401 {object} exceptions.ResponseError
// @Failure 403 {object} exceptions.ResponseError
// @Failure 404 {object} exceptions.ResponseError
// @Failure 422 {object} exceptions.ResponseError
// @Router /events/{id} [patch]
func (h *EventHandler) UpdateEvent(c echo.Context) error {
	userId, err := util.GetUserId(c)
	if err != nil {
		return err
	}
	id, err := util.GetParamsUUID(c, "id")
	if err != nil {
		return err
	}
	var body dto.UpdateEventDto
	if err := c.Bind(&body); err != nil {
		return exceptions.NewResponseError(exceptions.ErrBadRequest, err)
	}
	if err := c.Validate(&body); err != nil {
		return err
	}
	event, err := h.service.UpdateEvent(userId, id, body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, util.DataWrapper[*dto.EventResponseDto]{Data: event})
}

// @Summary Delete event
// @Description Deletes an event. Only the owner may delete it.
// @Tags Events
// @Security Bearer
// @Param id path string true "Event id (UUID)"
// @Success 204
// @Failure 401 {object} exceptions.ResponseError
// @Failure 403 {object} exceptions.ResponseError
// @Failure 404 {object} exceptions.ResponseError
// @Router /events/{id} [delete]
func (h *EventHandler) DeleteEvent(c echo.Context) error {
	userId, err := util.GetUserId(c)
	if err != nil {
		return err
	}
	id, err := util.GetParamsUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteEvent(userId, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
