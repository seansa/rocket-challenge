package repository

import (
	"fmt"
	"sort"
	"sync"
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
	db    map[string]T
	mutex sync.RWMutex
}

func NewRepository[T Storable]() Repository[T] {
	return &repository[T]{
		db: make(map[string]T),
	}
}

func (r *repository[T]) Get(key string) (T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	item, exists := r.db[key]
	if !exists {
		var zero T
		return zero, fmt.Errorf("key %s not found", key)
	}
	return item, nil
}

func (r *repository[T]) GetAll() ([]T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	items := make([]T, 0, len(r.db))

	for _, item := range r.db {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].GetKey() < items[j].GetKey()
	})

	return items, nil
}

func (r *repository[T]) Save(item T) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.db[item.GetKey()] = item
	return nil
}
