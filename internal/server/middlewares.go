package server

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jp-roisin/catch-and-go/internal/database/store"
	"github.com/labstack/echo/v4"
)

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
				session, err = s.db.CreateSession(ctx, uuid.New().String()) // TODO: get client defaults
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
