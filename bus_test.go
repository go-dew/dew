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
