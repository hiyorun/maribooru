package routes

import "maribooru/internal/handlers"

func (av *VersionOne) Administrative() {
	userHandler := handlers.NewUserHandler(av.db, av.cfg, av.log)
	permissionHandler := handlers.NewPermissionHandler(av.db, av.cfg, av.log)
	admin := av.api.Group("/admin", av.mw.AdminMiddleware())

	manage := admin.Group("/manage")
	manage.POST("", userHandler.CreateAdmin)
	manage.GET("", userHandler.GetAllAdmin)
	manage.PUT("/:id", userHandler.AssignAdmin)
	manage.DELETE("/:id", userHandler.RemoveAdmin)

	user := admin.Group("/user")
	user.PUT("/:id", userHandler.AdministrativeUserUpdate)
	user.GET("/permission/:id", permissionHandler.GetByUserID)
	user.PUT("/permission", permissionHandler.Set)
}
