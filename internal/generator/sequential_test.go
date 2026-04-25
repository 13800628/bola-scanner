package generator

import "testing"

func TestSequentialGenerator(t *testing.T) {
	gen := &SequentialGenerator{
		currentID: 10,
		endID:     11,
	}

	t.Run("NewSequentialGenerator", func(t *testing.T) {
		gen := NewSequentialGenerator(1, 3, "item-")

		val, hasNext := gen.Next()
		if val != "item-1" || !hasNext {
			t.Errorf("expected item-1, true; got %s, %v", val, hasNext)
		}
	})

	t.Run("Basic sequence", func(t *testing.T) {
		val, hasNext := gen.Next()
		if val != "10" || !hasNext {
			t.Errorf("1st Next() = %v, %v; 10, true", val, hasNext)
		}

		val, hasNext = gen.Next()
		if val != "11" || hasNext {
			t.Errorf("2nd Next() = %v, %v; want 11, false", val, hasNext)
		}
	})

	t.Run("With Prefix", func(t *testing.T) {
		gen := &SequentialGenerator{
			currentID: 1,
			endID:     2,
			prefix:    "user_",
		}

		val, _ := gen.Next()
		if val != "user_1" {
			t.Errorf("expected user_1, got %s", val)
		}

		val, _ = gen.Next()
		if val != "user_2" {
			t.Errorf("expected user_2, got %s", val)
		}
	})
}
