package routes

import "maribooru/internal/handlers"

func (av *VersionOne) Settings() {
	handler := handlers.NewSettingsHandler(av.db, av.cfg, av.log)
	settings := av.api.Group("/settings")
	settings.GET("", handler.Get)
}
