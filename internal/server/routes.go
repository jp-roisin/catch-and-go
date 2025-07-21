package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jp-roisin/catch-and-go/cmd/web"
	"github.com/jp-roisin/catch-and-go/cmd/web/components"
	"github.com/jp-roisin/catch-and-go/internal/database/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(s.AnonymousSessionMiddleware())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/", s.BaseWebHandler)
	e.GET("/health", s.healthHandler)
	e.PUT("/locale", s.UpdateLocale)

	e.GET("/lines/empty_state", s.LinesEmptyStateHandler)
	e.GET("/lines/picker", s.LinesPickerHandler)
	e.GET("/stops/picker/:lineId", s.StopsPickerHandler)

	e.POST("/dashboards", s.CreateDashboardHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) UpdateLocale(c echo.Context) error {
	ctx := c.Request().Context()

	locale := c.FormValue("locale")
	if locale != "fr" && locale != "nl" {
		return c.String(http.StatusBadRequest, "Invalid locale")
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	param := store.UpdateLocaleParams{
		ID:     session.ID,
		Locale: locale,
	}

	err := s.db.UpdateLocale(ctx, param)
	if err != nil {
		return err
	}

	// Refresh the page to apply the new locale
	c.Response().Header().Set("HX-Refresh", "true")
	return c.JSON(http.StatusNoContent, nil)
}

func (s *Server) BaseWebHandler(c echo.Context) error {
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	component := web.Base(session.Locale)
	err := component.Render(c.Request().Context(), c.Response())
	if err != nil {
		log.Printf("Error rendering in BaseWebHandler: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (s *Server) LinesPickerHandler(c echo.Context) error {
	ctx := c.Request().Context()
	var linesWithFallback []store.LineWithFallback
	lines, err := s.db.ListLinesByDirection(ctx, int(store.TowardsCity))
	for _, l := range lines {
		linesWithFallback = append(linesWithFallback, l.AddFallback())
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the lines info")
	}

	var sb strings.Builder
	if err := components.LinePicker(linesWithFallback).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the line pickers failed")
	}

	return c.HTML(http.StatusOK, sb.String())
}

func (s *Server) LinesEmptyStateHandler(c echo.Context) error {
	var sb strings.Builder
	if err := components.Empty_state().Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the empty state failed")
	}

	return c.HTML(http.StatusOK, sb.String())
}

func (s *Server) StopsPickerHandler(c echo.Context) error {
	ctx := c.Request().Context()
	lineID := c.Param("lineId")
	id, err := strconv.Atoi(lineID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid lineId: %q is not a number", lineID))
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	stops, err := s.db.ListStopsFromLine(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the stops info")
	}

	var props []store.Stop
	for _, s := range stops {
		translatedStop, err := s.Translate(session.Locale)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Something is wrong about this stop info: %v", s.Code))
		}
		props = append(props, translatedStop)
	}

	var sb strings.Builder
	if err := components.StopPicker(props).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the stops pickers failed")
	}

	return c.HTML(http.StatusOK, sb.String())
}

func (s *Server) CreateDashboardHandler(c echo.Context) error {
	ctx := c.Request().Context()
	stopId, err := strconv.Atoi(c.FormValue("stop_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid stop_id: must be an integer")
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	_, dbErr := s.db.CreateDashboard(ctx, store.CreatedashboardParams{
		SessionID: session.ID,
		StopID:    int64(stopId),
	})
	if dbErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't persist the dashboard")
	}

	var sb strings.Builder
	if err := components.Empty_state().Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the empty state failed")
	}

	return c.HTML(http.StatusCreated, sb.String())

}
