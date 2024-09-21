package routes

import (
	"maribooru/internal/structs"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (av *VersionOne) Heartbeat() {
	hb := av.api.Group("/heartbeat")
	hb.GET("/", OK)
	hb.GET("/admin-only", OK, av.mw.AdminMiddleware())
	hb.GET("/moderator", OK, av.mw.PermissionMiddleware(structs.Moderate))
	hb.GET("/approver", OK, av.mw.PermissionMiddleware(structs.Approve))
	hb.GET("/read-write", OK, av.mw.PermissionMiddleware(structs.Read|structs.Write))
	hb.GET("/write-only", OK, av.mw.PermissionMiddleware(structs.Write))
	hb.GET("/read-only", OK, av.mw.PermissionMiddleware(structs.Read))
}

func OK(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
