package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (av *VersionOne) Heartbeat() {
	hb := av.api.Group("/heartbeat")
	hb.GET("", OK)
}

func OK(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
