package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/keanutaufan/kvstored/api/controller"
)

func KeyValueRoutes(router *gin.Engine, keyValueController controller.KeyValueController) {
	routes := router.Group("/kv")
	{
		routes.GET("/:app_id", keyValueController.GetAll)
		routes.GET("/:app_id/:key", keyValueController.Get)
		routes.POST("/", keyValueController.Set)
		routes.PUT("/", keyValueController.Update)
		routes.DELETE("/:app_id/:key", keyValueController.Delete)
	}
}
