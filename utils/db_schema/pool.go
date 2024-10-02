package schema

import (
	"reflect"
	"sync"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TypePool defines the interface for a pool of new values
type TypePool interface {
	Get() any
	Put(any)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	// Thread-safe map of sync pools
	normalPool sync.Map

	// poolInitializer initializes or retrieves a reusable object pool for the specified
	// reflect.Type. It uses sync.Map to maintain a separate sync.Pool for each type, keyed by
	// reflect.Type. If a pool does not already exist for the given type, it creates a new sync.Pool
	// where each new object is an instance of the specified type created via reflection
	poolInitializer = func(reflectType reflect.Type) TypePool {
		v, _ := normalPool.LoadOrStore(reflectType, &sync.Pool{
			New: func() any {
				return reflect.New(reflectType).Interface()
			},
		})
		return v.(TypePool)
	}
)
