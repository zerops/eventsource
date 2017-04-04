package eventsource_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zerops/eventsource"
)

type Entity struct {
	Version   int
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EntityCreated struct {
	eventsource.Model
}

type EntityNameSet struct {
	eventsource.Model
	Name string
}

func (item *Entity) On(event eventsource.Event) bool {
	switch v := event.(type) {
	case *EntityCreated:
		item.Version = v.Model.Version
		item.ID = v.Model.ID
		item.CreatedAt = v.Model.At
		item.UpdatedAt = v.Model.At

	case *EntityNameSet:
		item.Version = v.Model.Version
		item.Name = v.Name
		item.UpdatedAt = v.Model.At

	default:
		return false
	}

	return true
}

func TestNew(t *testing.T) {
	repository := eventsource.New(&Entity{})
	aggregate := repository.New()
	assert.NotNil(t, aggregate)
	assert.Equal(t, &Entity{}, aggregate)
}

func TestRegistry(t *testing.T) {
	ctx := context.Background()
	id := "123"
	name := "Jones"

	t.Run("simple", func(t *testing.T) {
		registry := eventsource.New(&Entity{}, eventsource.WithDebug(os.Stdout))
		registry.Bind(EntityCreated{})
		registry.Bind(EntityNameSet{})

		// Test - Add an event to the store and verify we can recreate the object

		err := registry.Save(ctx,
			&EntityCreated{
				Model: eventsource.Model{ID: id, Version: 0, At: time.Unix(3, 0)},
			},
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 1, At: time.Unix(4, 0)},
				Name:  name,
			},
		)
		assert.Nil(t, err)

		v, err := registry.Load(ctx, id)
		assert.Nil(t, err, "expected successful load")
		fmt.Printf("%#v\n", v)

		org, ok := v.(*Entity)
		assert.True(t, ok)
		assert.Equal(t, id, org.ID, "expected restored id")
		assert.Equal(t, name, org.Name, "expected restored name")

		// Test - Update the org name and verify that the change is reflected in the loaded result

		updated := "Sarah"
		err = registry.Save(ctx, &EntityNameSet{
			Model: eventsource.Model{ID: id, Version: 2},
			Name:  updated,
		})
		assert.Nil(t, err)

		v, err = registry.Load(ctx, id)
		assert.Nil(t, err)

		org, ok = v.(*Entity)
		assert.True(t, ok)
		assert.Equal(t, id, org.ID)
		assert.Equal(t, updated, org.Name)
	})

	t.Run("with pointer prototype", func(t *testing.T) {
		registry := eventsource.New(&Entity{})
		registry.Bind(EntityCreated{})
		registry.Bind(EntityNameSet{})

		err := registry.Save(ctx,
			&EntityCreated{
				Model: eventsource.Model{ID: id, Version: 0, At: time.Unix(3, 0)},
			},
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 1, At: time.Unix(4, 0)},
				Name:  name,
			},
		)
		assert.Nil(t, err)

		v, err := registry.Load(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, name, v.(*Entity).Name)
	})

	t.Run("with pointer bind", func(t *testing.T) {
		registry := eventsource.New(&Entity{})
		registry.Bind(&EntityNameSet{})

		err := registry.Save(ctx,
			&EntityNameSet{
				Model: eventsource.Model{ID: id, Version: 0},
				Name:  name,
			},
		)
		assert.Nil(t, err)

		v, err := registry.Load(ctx, id)
		assert.Nil(t, err)
		assert.Equal(t, name, v.(*Entity).Name)
	})
}

func TestAt(t *testing.T) {
	ctx := context.Background()
	id := "123"

	registry := eventsource.New(&Entity{}, eventsource.WithDebug(os.Stdout))
	registry.Bind(EntityCreated{})
	err := registry.Save(ctx,
		&EntityCreated{
			Model: eventsource.Model{ID: id, Version: 1, At: time.Now()},
		},
	)
	assert.Nil(t, err)

	v, err := registry.Load(ctx, id)
	assert.Nil(t, err)

	org := v.(*Entity)
	assert.NotZero(t, org.CreatedAt)
	assert.NotZero(t, org.UpdatedAt)
}
