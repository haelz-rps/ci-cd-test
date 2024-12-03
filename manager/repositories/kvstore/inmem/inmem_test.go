package inmem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ubix/dripper/manager/repositories/kvstore"
)

func TestGet(t *testing.T) {
	uuidKey := uuid.New()
	missingKey := uuid.New()

	testCases := []struct {
		name          string
		setup         func(store *inmemStore)
		key           uuid.UUID
		expectedValue interface{}
		expectError   bool
		expectedErr   error
	}{
		{
			name: "KeyExists",
			setup: func(store *inmemStore) {
				store.Set(context.Background(), uuidKey, "bar")
			},
			key:           uuidKey,
			expectedValue: "bar",
			expectError:   false,
		},
		{
			name: "KeyDoesNotExist",
			setup: func(store *inmemStore) {
				// No setup needed; missing key does not exist
			},
			key:         missingKey,
			expectError: true,
			expectedErr: kvstore.ErrKeyNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Initialize a new inmemStore for each test case to ensure isolation
			store := NewInmemStore()

			// Setup the store as per the test case
			if tc.setup != nil {
				tc.setup(store)
			}

			// Call the Get method
			value, err := store.Get(context.Background(), tc.key)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				assert.Equal(t, tc.expectedValue, value)
			}
		})
	}
}

func TestSet(t *testing.T) {
	uuidKey := uuid.New()
	uuidKey2 := uuid.New()
	uuidKey3 := uuid.New()

	testCases := []struct {
		name          string
		setup         func(store *inmemStore)
		key           uuid.UUID
		value         interface{}
		expectedValue interface{}
	}{
		{
			name: "SetNewKey",
			setup: func(store *inmemStore) {
				// No pre-existing keys
			},
			key:           uuidKey,
			value:         "bar",
			expectedValue: "bar",
		},
		{
			name: "OverwriteExistingKey",
			setup: func(store *inmemStore) {
				store.Set(context.Background(), uuidKey, "bar")
			},
			key:           uuidKey,
			value:         "baz",
			expectedValue: "baz",
		},
		{
			name: "SetMultipleKeys",
			setup: func(store *inmemStore) {
				store.Set(context.Background(), uuidKey, "value1")
				store.Set(context.Background(), uuidKey2, 42)
			},
			key:           uuidKey3,
			value:         struct{ Name string }{Name: "Alice"},
			expectedValue: struct{ Name string }{Name: "Alice"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Initialize a new inmemStore for each test case to ensure isolation
			store := NewInmemStore()

			// Setup the store as per the test case
			if tc.setup != nil {
				tc.setup(store)
			}

			// Call the Set method
			err := store.Set(context.Background(), tc.key, tc.value)

			// Since Set always returns nil, assert no error
			assert.NoError(t, err, "Set should not return an error")

			// Directly access the internal store map to verify
			store.m.Lock()
			defer store.m.Unlock()

			actualValue, _ := store.store[tc.key]

			if tc.expectedValue == nil {
				// If expected value is nil, check for existence
				assert.Nil(t, actualValue, "Expected value to be nil")
			} else {
				assert.Equal(t, tc.expectedValue, actualValue, "Retrieved value does not match expected")
			}
		})
	}
}
