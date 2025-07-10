package server

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/jp-roisin/catch-and-go/cmd/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: `cookie:"token"`,
		Validator: func(key string, c echo.Context) (bool, error) {
			ctx := c.Request().Context()

			// TODO: Cache tokens in memory to avoid hitting the DB on every request.
			// A simple in-memory map might be enough, or consider using an LRU cache:
			// https://github.com/hashicorp/golang-lru
			_, err := s.db.GetSession(ctx, key)
			if err != nil {
				return false, c.String(http.StatusUnauthorized, "Session token is not valid")
			}
			return true, nil
		},
		ErrorHandler: func(err error, c echo.Context) error {
			ctx := c.Request().Context()

			session, err := s.db.CreateSession(ctx, uuid.New().String())
			if err != nil {
				return c.String(http.StatusInternalServerError, "Couldn't create a session token")
			}

			c.SetCookie(
				&http.Cookie{
					Name:     "token",
					Value:    session.ID,
					Path:     "/",
					HttpOnly: true,
					Expires:  time.Now().Add(365 * 24 * time.Hour),
				},
			)

			return nil
		},
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	e.GET("/", s.HelloWorldHandler)

	e.GET("/health", s.healthHandler)

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
