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
	"github.com/jp-roisin/catch-and-go/internal/externalapi"
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

	e.Static("/assets", "cmd/web/assets")

	e.GET("/", s.BaseWebHandler)
	e.GET("/main", s.MainHandler)
	e.GET("/health", s.healthHandler)

	e.GET("/sessions", s.GetSessionHandler)
	e.PUT("/sessions/locale", s.UpdateLocaleHandler)
	e.PUT("/sessions/theme", s.UpdateThemeHandler)

	e.GET("/lines/empty_state", s.LinesEmptyStateHandler)
	e.GET("/lines/picker", s.LinesPickerHandler)
	e.GET("/stops/picker/:lineId", s.StopsPickerHandler)

	e.GET("/dashboards", s.GetDashboardsHandler)
	e.GET("/dashboards/:dashboardId", s.GetDashboardContentHandler)
	e.POST("/dashboards", s.CreateDashboardHandler)
	e.DELETE("/dashboards/:dashboardId", s.DeleteDashboardHandler)

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

func (s *Server) GetSessionHandler(c echo.Context) error {
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	var sb strings.Builder
	if err := components.Header(session).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the line pickers failed")
	}

	return c.HTML(http.StatusOK, sb.String())
}

func (s *Server) UpdateLocaleHandler(c echo.Context) error {
	ctx := c.Request().Context()
	newLocale := c.FormValue("locale")
	if newLocale != "fr" && newLocale != "nl" {
		return echo.NewHTTPError(http.StatusBadRequest, "Locale is not valid")
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	param := store.UpdateLocaleParams{
		ID:     session.ID,
		Locale: newLocale,
	}

	err := s.db.UpdateLocale(ctx, param)
	if err != nil {
		return err
	}

	component := components.Main()
	renderingErr := component.Render(c.Request().Context(), c.Response())
	if renderingErr != nil {
		log.Printf("Error rendering in main content: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, renderingErr.Error())
	}
	return nil
}

func (s *Server) UpdateThemeHandler(c echo.Context) error {
	ctx := c.Request().Context()
	newTheme := c.FormValue("theme")
	if newTheme != "light" && newTheme != "dark" {
		return echo.NewHTTPError(http.StatusBadRequest, "Theme is not valid")
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	param := store.UpdateThemeParams{
		ID:    session.ID,
		Theme: newTheme,
	}

	err := s.db.UpdateTheme(ctx, param)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	var sb strings.Builder
	if err := components.ThemeSwitch(param.Theme).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the ThemeSwitch failed")
	}

	return c.HTML(http.StatusOK, sb.String())
}

func (s *Server) MainHandler(c echo.Context) error {
	component := components.Main()
	err := component.Render(c.Request().Context(), c.Response())
	if err != nil {
		log.Printf("Error rendering in main content: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (s *Server) BaseWebHandler(c echo.Context) error {
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	component := web.Base(session.Theme)
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
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	var sb strings.Builder
	if err := components.EmptyState(session.Theme).Render(c.Request().Context(), &sb); err != nil {
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
	if err := components.Main().Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the empty state failed")
	}

	return c.HTML(http.StatusCreated, sb.String())
}

func (s *Server) GetDashboardsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	dashboards, err := s.db.ListDashboardsFromSession(ctx, session.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the dashboard")
	}

	var translatedDashboards []store.ListDashboardsFromSessionRow
	for _, d := range dashboards {
		translatedDashboard, err := d.Translate(session.Locale)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Something is wrong about this stop info: %v", d.DashboardID))
		}
		translatedDashboards = append(translatedDashboards, translatedDashboard)

	}

	var sb strings.Builder
	if err := components.Dashboard(translatedDashboards).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the empty state failed")
	}

	return c.HTML(http.StatusCreated, sb.String())
}

func (s *Server) DeleteDashboardHandler(c echo.Context) error {
	ctx := c.Request().Context()
	param := c.Param("dashboardId")
	dashboardId, err := strconv.Atoi(param)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid dashboardId: %q is not a number", dashboardId))
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	err = s.db.DeleteDashboard(ctx, store.DeleteDashboardParams{
		ID:        int64(dashboardId),
		SessionID: session.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't delete the dashboard")
	}

	return c.Blob(http.StatusOK, "text/html", []byte(""))
}

func (s *Server) GetDashboardContentHandler(c echo.Context) error {
	ctx := c.Request().Context()
	param := c.Param("dashboardId")
	dashboardId, err := strconv.Atoi(param)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid dashboardId: %q is not a number", dashboardId))
	}

	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	d, err := s.db.GetDashboardByIdWithStopInfo(ctx, store.GetDashboardByIdWithStopInfoParams{
		ID:        int64(dashboardId),
		SessionID: session.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the dashboard with stop info")
	}

	res, err := externalapi.GetWaitingTimeForStop(d.StopCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	lineId := res.WaitingTimes[0].LineID
	line, err := s.db.GetLine(ctx, store.GetLineParams{
		Code:      lineId,
		Direction: 0, // We're only looking for the metadata which are the same in both directions
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the line info")
	}

	var sb strings.Builder
	if err := components.DashboardContent(components.DashboardContentProps{
		WaitingTimes: res.WaitingTimes,
		Line:         line,
		Locale:       session.Locale,
	}).Render(c.Request().Context(), &sb); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Rendering of the empty state failed")
	}

	return c.HTML(http.StatusCreated, sb.String())
}
