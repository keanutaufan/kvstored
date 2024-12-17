package dto

type KeyValueGetRequest struct {
	AppID string `json:"app_id" binding:"required"`
	Key   string `json:"key" binding:"required"`
}

type KeyValueSetRequest struct {
	AppID string `json:"app_id" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type KeyValueUpdateRequest struct {
	AppID string `json:"app_id" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}
