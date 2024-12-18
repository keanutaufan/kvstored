package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/gocql/gocql"
	"github.com/keanutaufan/kvstored/api/db"
	"github.com/keanutaufan/kvstored/api/entity"
)

type KeyValueRepository interface {
	GetAll(ctx context.Context, appID string) ([]entity.KeyValue, error)
	Set(ctx context.Context, keyValue entity.KeyValue) error
	Get(ctx context.Context, appID, key string) (entity.KeyValue, error)
	Update(ctx context.Context, keyValue entity.KeyValue) error
	Delete(ctx context.Context, appID, key string) error
}

type keyValueRepository struct {
	client *db.CassandraClient
}

func NewKeyValueRepository(client *db.CassandraClient) KeyValueRepository {
	return &keyValueRepository{client: client}
}

// Add to repository/key_value.go
func (r *keyValueRepository) GetAll(ctx context.Context, appID string) ([]entity.KeyValue, error) {
	var keyValues []entity.KeyValue
	iter := r.client.Session.Query(`
        SELECT app_id, key, value, created_at 
        FROM kv_store_app.key_values 
        WHERE app_id = ?
    `, appID).WithContext(ctx).Iter()

	var kv entity.KeyValue
	for iter.Scan(&kv.AppID, &kv.Key, &kv.Value, &kv.CreatedAt) {
		keyValues = append(keyValues, kv)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return keyValues, nil
}

func (r *keyValueRepository) Set(ctx context.Context, keyValue entity.KeyValue) error {
	if strings.TrimSpace(keyValue.AppID) == "" || strings.TrimSpace(keyValue.Key) == "" {
		return errors.New("app_id and key cannot be empty")
	}

	return r.client.Session.Query(`
        INSERT INTO kv_store_app.key_values (app_id, key, value, created_at) 
        VALUES (?, ?, ?, ?)
    `, keyValue.AppID, keyValue.Key, keyValue.Value, keyValue.CreatedAt).WithContext(ctx).Exec()
}

func (r *keyValueRepository) Get(ctx context.Context, appID, key string) (entity.KeyValue, error) {
	var keyValue entity.KeyValue
	err := r.client.Session.Query(`
		SELECT app_id, key, value, created_at FROM kv_store_app.key_values 
		WHERE app_id = ? AND key = ?
	`, appID, key).WithContext(ctx).Scan(&keyValue.AppID, &keyValue.Key, &keyValue.Value, &keyValue.CreatedAt)

	if err == gocql.ErrNotFound {
		return entity.KeyValue{}, errors.New("key not found for the given app")
	}
	return keyValue, err
}

func (r *keyValueRepository) Update(ctx context.Context, keyValue entity.KeyValue) error {
	if strings.TrimSpace(keyValue.AppID) == "" || strings.TrimSpace(keyValue.Value) == "" {
		return errors.New("app_id and key cannot be empty")
	}

	var existing string
	err := r.client.Session.Query(`
        SELECT value FROM kv_store_app.key_values 
        WHERE app_id = ? AND key = ?
    `, keyValue.AppID, keyValue.Key).WithContext(ctx).Scan(&existing)

	if err == gocql.ErrNotFound {
		return errors.New("key not found for the given app")
	} else if err != nil {
		return err
	}

	return r.client.Session.Query(`
        UPDATE kv_store_app.key_values 
        SET value = ?
        WHERE app_id = ? AND key = ?
    `, keyValue.Value, keyValue.AppID, keyValue.Key).WithContext(ctx).Exec()
}

func (r *keyValueRepository) Delete(ctx context.Context, appID, key string) error {
	return r.client.Session.Query(`
        DELETE FROM kv_store_app.key_values 
        WHERE app_id = ? AND key = ?
    `, appID, key).WithContext(ctx).Exec()
}
