package server

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jp-roisin/catch-and-go/cmd/web"
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

	e.GET("/", web.BaseWebHandler)
	e.GET("/health", s.healthHandler)
	e.PUT("/locale", s.UpdateLocale)

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

func (s *Server) AnonymousSessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			cookie, err := c.Cookie("token")
			var session store.Session

			if err == nil {
				session, err = s.db.GetSession(ctx, cookie.Value)

				if err != nil {
					// Token is invalid or expired — reject the request
					// TODO clear previous + create a new token
					log.Printf("Invalid session token: %s", cookie.Value)
					return c.String(http.StatusUnauthorized, "Invalid session token")
				}
			} else {
				// Create a new anonymous session
				session, err = s.db.CreateSession(ctx, uuid.New().String())
				if err != nil {
					return c.String(http.StatusInternalServerError, "Couldn't create session")
				}

				c.SetCookie(&http.Cookie{
					Name:     "token",
					Value:    session.ID,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
					Expires:  time.Now().Add(365 * 24 * time.Hour),
				})
			}

			c.Set("session", &session)
			log.Printf("Active session: %s", session.ID)

			// Continue the request
			return next(c)
		}
	}
}
