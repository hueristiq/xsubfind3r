//nolint:testpackage // White-box tests pin the KeysManager rotation internals.
package github

import (
	"sync"
	"testing"
	"time"
)

func TestKeysManagerRotation(t *testing.T) {
	t.Parallel()

	manager := NewKeyManager([]string{"a", "b", "c"})

	if key := manager.GetCurrentKey(); key.Value != "a" {
		t.Fatalf("first key = %q, want %q", key.Value, "a")
	}

	if key := manager.GetCurrentKey(); key.Value != "b" {
		t.Fatalf("second key = %q, want %q", key.Value, "b")
	}

	if key := manager.GetCurrentKey(); key.Value != "c" {
		t.Fatalf("third key = %q, want %q", key.Value, "c")
	}

	if key := manager.GetCurrentKey(); key.Value != "a" {
		t.Fatalf("wrapped key = %q, want %q", key.Value, "a")
	}
}

func TestKeysManagerSetCurrentKeyExceeded(t *testing.T) {
	t.Parallel()

	manager := NewKeyManager([]string{"a", "b"})

	_ = manager.GetCurrentKey()

	manager.SetCurrentKeyExceeded(60)

	if manager.pool[0].RetryAfter != 60 {
		t.Fatalf("pool[0].RetryAfter = %d, want 60 (the key just used)", manager.pool[0].RetryAfter)
	}

	if manager.pool[1].RetryAfter != 0 {
		t.Fatalf("pool[1].RetryAfter = %d, want 0 (the next key must stay usable)", manager.pool[1].RetryAfter)
	}

	if key := manager.GetCurrentKey(); key.Value != "b" {
		t.Fatalf("next key = %q, want %q", key.Value, "b")
	}

	manager.SetCurrentKeyExceeded(60)

	if manager.pool[1].RetryAfter != 60 {
		t.Fatalf("pool[1].RetryAfter = %d, want 60 (the key just used)", manager.pool[1].RetryAfter)
	}
}

func TestKeysManagerCooldownExpiry(t *testing.T) {
	t.Parallel()

	manager := NewKeyManager([]string{"a"})

	_ = manager.GetCurrentKey()

	manager.SetCurrentKeyExceeded(1)

	manager.pool[0].ExceededTime = time.Now().Add(-2 * time.Second)

	key := manager.GetCurrentKey()

	if key.RetryAfter != 0 {
		t.Fatalf("RetryAfter = %d, want 0 after cooldown elapsed", key.RetryAfter)
	}
}

func TestKeysManagerConcurrent(t *testing.T) {
	t.Parallel()

	manager := NewKeyManager([]string{"a", "b", "c", "d"})

	wg := &sync.WaitGroup{}

	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 50 {
				_ = manager.GetCurrentKey()
				manager.SetCurrentKeyExceeded(1)
			}
		}()
	}

	wg.Wait()
}
