package routes

import (
	"maribooru/internal/setting"
)

func (av *VersionOne) Settings() {
	handler := setting.NewHandler(av.db, av.cfg, av.log)
	settings := av.api.Group("/settings")
	settings.GET("", handler.Get)
}
