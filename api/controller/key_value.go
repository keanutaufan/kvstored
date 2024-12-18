package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keanutaufan/kvstored/api/dto"
	"github.com/keanutaufan/kvstored/api/entity"
	"github.com/keanutaufan/kvstored/api/realtime"
	"github.com/keanutaufan/kvstored/api/service"
)

type KeyValueController interface {
	GetAll(ctx *gin.Context)
	Set(ctx *gin.Context)
	Get(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type keyValueController struct {
	keyValueService service.KeyValueService
	kafkaService    *realtime.KafkaService
}

func NewKeyValueController(keyValueService service.KeyValueService, kafkaService *realtime.KafkaService) *keyValueController {
	return &keyValueController{
		keyValueService: keyValueService,
		kafkaService:    kafkaService,
	}
}

func (c *keyValueController) GetAll(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	keyValues, err := c.keyValueService.GetAll(ctx.Request.Context(), appID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(keyValues) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no keys found for the given app"})
		return
	}

	ctx.JSON(http.StatusOK, keyValues)
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

	if err := c.kafkaService.PublishKeyChange("set", keyValue.AppID, keyValue.Key, &keyValue); err != nil {
		log.Printf("Error publishing Kafka message: %v", err)
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

	if err := c.kafkaService.PublishKeyChange("update", keyValue.AppID, keyValue.Key, &keyValue); err != nil {
		log.Printf("Error publishing Kafka message: %v", err)
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

	if err := c.kafkaService.PublishKeyChange("delete", appID, key, nil); err != nil {
		log.Printf("Error publishing Kafka message: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
