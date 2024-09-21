package routes

import "maribooru/internal/handlers"

func (av *VersionOne) Administrative() {
	userHandler := handlers.NewUserHandler(av.db, av.cfg, av.log)
	permissionHandler := handlers.NewPermissionHandler(av.db, av.cfg, av.log)
	admin := av.api.Group("/admin", av.mw.AdminMiddleware())

	user := admin.Group("/user")
	user.GET("/permission/:id", permissionHandler.GetByUserID)
	user.POST("/permission", permissionHandler.Set)
	user.POST("/assign-admin/:id", userHandler.AssignAdmin)
	user.POST("/create-admin", userHandler.CreateAdmin)
}
