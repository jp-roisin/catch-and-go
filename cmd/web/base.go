package web

import (
	"log"
	"net/http"

	"github.com/jp-roisin/catch-and-go/internal/database/store"
	"github.com/labstack/echo/v4"
)

func BaseWebHandler(c echo.Context) error {
	session, ok := c.Get("session").(*store.Session)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Couldn't retreive the session token")
	}

	component := Base(session.Locale)
	err := component.Render(c.Request().Context(), c.Response())
	if err != nil {
		log.Printf("Error rendering in BaseWebHandler: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
