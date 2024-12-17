package service

import (
	"context"
	"errors"

	"github.com/keanutaufan/kvstored/api/entity"
	"github.com/keanutaufan/kvstored/api/repository"
)

type KeyValueService interface {
	Set(ctx context.Context, keyValue entity.KeyValue) error
	Get(ctx context.Context, appID, key string) (entity.KeyValue, error)
	Update(ctx context.Context, keyValue entity.KeyValue) error
	Delete(ctx context.Context, appID, key string) error
}

type keyValueService struct {
	kvRepository repository.KeyValueRepository
}

func NewKeyValueService(kvRepository repository.KeyValueRepository) *keyValueService {
	return &keyValueService{kvRepository: kvRepository}
}

func (s *keyValueService) Set(ctx context.Context, keyValue entity.KeyValue) error {
	if keyValue.Key == "" {
		return errors.New("key cannot be empty")
	}
	if keyValue.Value == "" {
		return errors.New("value cannot be empty")
	}
	return s.kvRepository.Set(ctx, keyValue)
}

func (s *keyValueService) Get(ctx context.Context, appID, key string) (entity.KeyValue, error) {
	value, err := s.kvRepository.Get(ctx, appID, key)
	if err != nil {
		return entity.KeyValue{}, err
	}
	return value, nil
}

func (s *keyValueService) Update(ctx context.Context, keyValue entity.KeyValue) error {
	if keyValue.Key == "" {
		return errors.New("key cannot be empty")
	}
	if keyValue.Value == "" {
		return errors.New("value cannot be empty")
	}
	return s.kvRepository.Update(ctx, keyValue)
}

func (s *keyValueService) Delete(ctx context.Context, appID, key string) error {
	return s.kvRepository.Delete(ctx, appID, key)
}
