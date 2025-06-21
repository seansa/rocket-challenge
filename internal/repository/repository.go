package repository

import (
	"fmt"
)

type Storable interface {
	GetKey() string
}

type Repository[T Storable] interface {
	Get(key string) (T, error)
	GetAll() ([]T, error)
	Save(item T) error
}

type repository[T Storable] struct {
	db map[string]T
}

func NewRepository[T Storable]() Repository[T] {
	return &repository[T]{
		db: make(map[string]T),
	}
}

func (r *repository[T]) Get(key string) (T, error) {
	item, exists := r.db[key]
	if !exists {
		var zero T
		return zero, fmt.Errorf("key %s not found", key)
	}
	return item, nil
}

func (r *repository[T]) GetAll() ([]T, error) {
	var items []T

	for _, item := range r.db {
		items = append(items, item)
	}

	return items, nil
}

func (r *repository[T]) Save(item T) error {
	r.db[item.GetKey()] = item
	return nil
}
