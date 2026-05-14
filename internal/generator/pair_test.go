package generator

import "testing"

func TestPairGenerator(t *testing.T) {

	t.Run("Returns all pairs in oeder", func(t *testing.T) {
		gen := NewPairGenerator([]string{"1", "2", "3"})

		expected := [][2]string{
			{"1", "2"},
			{"1", "3"},
			{"2", "3"},
		}

		for i, exp := range expected {
			victim, attacker, hasNext := gen.Next()
			if victim != exp[0] || attacker != exp[1] {
				t.Errorf("pair %d expected (%s, %s), got (%s, %s)", i, exp[0], exp[1], victim, attacker)
			}
			isLast := i == len(expected)-1
			if hasNext == isLast {
				t.Errorf("pair %d: expected hasNext=%v, got %v", i, !isLast, hasNext)
			}
		}
	})

	t.Run("Two IDs returns single pair", func(t *testing.T) {
		gen := NewPairGenerator([]string{"1", "2"})
		victim, attacker, hasNext := gen.Next()
		if victim != "1" || attacker != "2" || hasNext {
			t.Errorf("expected (1, 2, false), got (%s, %s, %v)", victim, attacker, hasNext)
		}
	})

	t.Run("Single ID returns no pairs", func(t *testing.T) {
		gen := NewPairGenerator([]string{"1"})
		victim, attacker, hasNext := gen.Next()
		if victim != "" || attacker != "" || hasNext {
			t.Errorf("expected empty pair, got (%s, %s, %v)", victim, attacker, hasNext)
		}
	})

	t.Run("Empty IDs returns no paris", func(t *testing.T) {
		gen := NewPairGenerator([]string{})
		victim, attacker, hasNext := gen.Next()
		if victim != "" || attacker != "" || hasNext {
			t.Errorf("expected empty pair, got (%s, %s, %v)", victim, attacker, hasNext)
		}
	})

	t.Run("Exhausted generator returns empty", func(t *testing.T) {
		gen := NewPairGenerator([]string{"1", "2"})
		gen.Next()
		victim, attacker, hasNext := gen.Next()
		if victim != "" || attacker != "" || hasNext {
			t.Errorf("expected empty after exhausted, got (%s, %s, %v)", victim, attacker, hasNext)
		}
	})
}
