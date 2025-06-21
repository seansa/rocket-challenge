package repository

import (
	"testing"
	"time"

	"github.com/seansa/rocket-challenge/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestSaveAndGet tests saving and retrieving a single item.
func TestSaveAndGet(t *testing.T) {
	repo := NewRepository[model.Rocket]()
	item := model.NewRocket("item-test-1")
	item.Type = "Falcon-Heavy"
	item.Speed = 1000
	item.Mission = "Mars"
	item.MessageNumber = 10
	item.MessageTime = time.Now()

	err := repo.Save(item)
	assert.NoError(t, err)

	retrievedItem, err := repo.Get("item-test-1")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedItem)
	assert.Equal(t, item.Channel, retrievedItem.Channel)
	assert.Equal(t, item.Type, retrievedItem.Type)
	assert.Equal(t, item.Speed, retrievedItem.Speed)
	assert.Equal(t, item.Mission, retrievedItem.Mission)
	assert.Equal(t, item.MessageNumber, retrievedItem.MessageNumber)
	assert.Equal(t, item.MessageTime.Unix(), retrievedItem.MessageTime.Unix())
}

// TestGet_NotFound tests retrieving a non-existent item.
func TestGet_NotFound(t *testing.T) {
	repo := NewRepository[model.Rocket]()
	_, err := repo.Get("non-existent-item")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestGetAll tests retrieving all items and sorting.
func TestGetAll(t *testing.T) {
	repo := NewRepository[model.Rocket]()

	itemA := model.NewRocket("item-B")
	itemA.Speed = 200

	itemB := model.NewRocket("item-A")
	itemB.Speed = 100

	itemC := model.NewRocket("item-C")
	itemC.Speed = 300

	_ = repo.Save(itemA)
	_ = repo.Save(itemB)
	_ = repo.Save(itemC)

	allItems, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allItems, 3)

	// Verify sorting by key
	assert.Equal(t, "item-A", allItems[0].GetKey())
	assert.Equal(t, "item-B", allItems[1].GetKey())
	assert.Equal(t, "item-C", allItems[2].GetKey())
}

// TestSave_UpdateExisting tests updating an existing item.
func TestSave_UpdateExisting(t *testing.T) {
	repo := NewRepository[model.Rocket]()
	item := model.NewRocket("item-update")
	item.Speed = 100

	_ = repo.Save(item)

	item.Speed = 500
	item.Mission = "New Mission"
	_ = repo.Save(item)

	updatedItem, err := repo.Get("item-update")
	assert.NoError(t, err)
	assert.Equal(t, 500, updatedItem.Speed)
	assert.Equal(t, "New Mission", updatedItem.Mission)
}

// TestGetAll_Empty tests retrieving all item when none exist.
func TestGetAll_Empty(t *testing.T) {
	repo := NewRepository[model.Rocket]()
	allItems, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allItems, 0)
	assert.NotNil(t, allItems)
}
