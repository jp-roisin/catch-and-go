package server

import (
	"log"
	"net/http"
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

	e.GET("/lines/picker", s.LinesPickerHandler)

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
