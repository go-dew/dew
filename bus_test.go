package dew

import (
	"context"
	"testing"
)

func TestMustFromContext(t *testing.T) {
	t.Run("Panic if bus is not found in context", func(t *testing.T) {
		defer func() {
			// recover from panic
			if r := recover(); r != nil {
				// check if the panic message is the expected one
				if r != "bus not found in context" {
					t.Errorf("expected panic message: bus not found in context, got: %v", r)
				}
			} else {
				t.Error("expected panic, got none")
			}
		}()
		ctx := context.Background()
		MustFromContext(ctx)
	})
}

func TestNewContext(t *testing.T) {
	t.Run("Return a new context with the given bus", func(t *testing.T) {
		bus := New()
		ctx := NewContext(context.Background(), bus)
		if ctx == nil {
			t.Error("expected context, got nil")
		}
		// check if the bus is in the context
		b, ok := FromContext(ctx)
		if !ok {
			t.Error("expected bus in context, got none")
		}
		if b != bus {
			t.Errorf("expected bus: %v, got: %v", bus, b)
		}
	})
}
