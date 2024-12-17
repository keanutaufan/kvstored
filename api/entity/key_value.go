package entity

import "time"

type KeyValue struct {
	AppID     string    `json:"app_id" binding:"required"`
	Key       string    `json:"key" binding:"required"`
	Value     string    `json:"value" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}
