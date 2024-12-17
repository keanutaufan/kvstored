package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keanutaufan/kvstored/api/dto"
	"github.com/keanutaufan/kvstored/api/entity"
	"github.com/keanutaufan/kvstored/api/service"
)

type KeyValueController interface {
	Set(ctx *gin.Context)
	Get(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type keyValueController struct {
	keyValueService service.KeyValueService
}

func NewKeyValueController(keyValueService service.KeyValueService) *keyValueController {
	return &keyValueController{keyValueService: keyValueService}
}

func (c *keyValueController) Set(ctx *gin.Context) {
	var req dto.KeyValueSetRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyValue := entity.KeyValue{
		AppID:     req.AppID,
		Key:       req.Key,
		Value:     req.Value,
		CreatedAt: time.Now(),
	}

	if err := c.keyValueService.Set(ctx.Request.Context(), keyValue); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (c *keyValueController) Get(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	key := ctx.Param("key")

	value, err := c.keyValueService.Get(ctx.Request.Context(), appID, key)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, value)
}

func (c *keyValueController) Update(ctx *gin.Context) {
	var req dto.KeyValueUpdateRequest

	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keyValue := entity.KeyValue{
		AppID: req.AppID,
		Key:   req.Key,
		Value: req.Value,
	}

	if err := c.keyValueService.Update(ctx.Request.Context(), keyValue); err != nil {
		if err.Error() == "key not found for the given app" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (c *keyValueController) Delete(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	key := ctx.Param("key")

	if err := c.keyValueService.Delete(ctx.Request.Context(), appID, key); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
