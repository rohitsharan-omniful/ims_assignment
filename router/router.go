package router

import (
	"github.com/gin-gonic/gin"
	"github.com/omniful/ims_rohit/handlers"
)

// SetupRouter configures all routes for the application
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Hub routes
		hubRoutes := v1.Group("/hubs")
		{
			hubRoutes.POST("/", handlers.CreateHubHandler)
			hubRoutes.GET("/", handlers.ListHubsHandler)
			hubRoutes.GET("/:id", handlers.GetHubHandler)
			hubRoutes.PUT("/:id", handlers.UpdateHubHandler)
			hubRoutes.DELETE("/:id", handlers.DeleteHubHandler)
		}

		// SKU routes
		skuRoutes := v1.Group("/skus")
		{
			skuRoutes.POST("/", handlers.CreateSKUHandler)
			skuRoutes.GET("/", handlers.ListSKUsHandler)
			skuRoutes.GET("/:id", handlers.GetSKUHandler)
			skuRoutes.PUT("/:id", handlers.UpdateSKUHandler)
			skuRoutes.DELETE("/:id", handlers.DeleteSKUHandler)
			skuRoutes.POST("/validate", handlers.CheckSKUsExistenceHandler)
		}

		// Inventory routes
		inventoryRoutes := v1.Group("/inventory")
		{
			inventoryRoutes.POST("/upsert", handlers.UpsertInventoryHandler)
			inventoryRoutes.POST("/view", handlers.ViewInventoryHandler)
		}
	}

	return r
}
