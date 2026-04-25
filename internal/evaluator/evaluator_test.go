package evaluator

import "testing"

func TestEvaluateStatus(t *testing.T) {
	ev := &Evaluator{
		StatusWeight: 30,
	}

	t.Run("Status code match gives full score", func(t *testing.T) {
		score := ev.evaluateStatus(200, 200)
		expected := 30

		if score != expected {
			t.Errorf("expected 30, but got %d", score)
		}
	})

	t.Run("Status code mismatch gives zero", func(t *testing.T) {
		score := ev.evaluateStatus(200, 403)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}

// JSON Test
func TestEvaluator_EvaluateStructure(t *testing.T) {
	ev := &Evaluator{
		StatusWeight:    30,
		StructureWeight: 30,
	}

	t.Run("JSON keys match perfectly", func(t *testing.T) {
		victimJSON := `{"id": 1, "name": "attacker", "email": "v@example.com"}`
		attackerJSON := `{"id": 100, "name": "attacker", "email": "a@example.com"}`

		score := ev.evaluateStructure(victimJSON, attackerJSON)
		if score != 30 {
			t.Errorf("expected 30, but got %d", score)
		}
	})

	t.Run("JSON keys mismatch", func(t *testing.T) {
		victimJSON := `{"id": 1, "name": "victim"}`
		attackerJSON := `{"error": "not found"}` // 全く違う構造

		score := ev.evaluateStructure(victimJSON, attackerJSON)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}

func TestEvaluator_EvaluateKeyWords(t *testing.T) {
	ev := &Evaluator{
		KeywordWeight:   25,
		StatusWeight:    0,
		StructureWeight: 0,
		SizeWeight:      0,
	}

	t.Run("Response contains sensitive keywords", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{}`}
		attacker := ResponseData{
			StatusCode: 200,
			Body:       `{"id": 100, "name": "Isayama", "note": "secret info"}`,
		}
		keywords := []string{"Isayama", "secret"}

		// ここを実装クラスで実装
		score := ev.FullEvaluation(victim, attacker, keywords)
		if score != 25 {
			t.Errorf("expected 25, but got %d", score)
		}
	})

	t.Run("Response does not contain keywords", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{}`}
		attacker := ResponseData{
			StatusCode: 200,
			Body:       `{"id": 101, "name": "Unknown", "note": "public"}`,
		}
		keywords := []string{"Isayama"}

		score := ev.FullEvaluation(victim, attacker, keywords)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}

func TestEvaluator_FullEvaluation(t *testing.T) {
	ev := &Evaluator{
		StatusWeight:    30,
		StructureWeight: 30,
		KeywordWeight:   25,
		SizeWeight:      15,
	}

	t.Run("Perfect Match - 100 points", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		keywords := []string{"victim"}

		score := ev.FullEvaluation(victim, attacker, keywords)

		if score != 100 {
			t.Errorf("expected 100, but got %d", score)
		}
	})

	t.Run("High probably BOLA (All criteria met)", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		keywords := []string{"victim"}

		score := ev.FullEvaluation(victim, attacker, keywords)

		if score != 100 {
			t.Errorf("expected 100, but got %d", score)
		}
	})

	t.Run("Low probably (Only status matches)", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim", "email": "v@example.com"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"status": "error", "message": "unauthorized access detected"}`}
		keywords := []string{"victim"}

		score := ev.FullEvaluation(victim, attacker, keywords)

		if score != 30 {
			t.Errorf("expected 30, but got %d", score)
		}
	})
}

func TestEvaluator_EvaluateSize(t *testing.T) {
	ev := &Evaluator{
		SizeWeight: 15,
	}

	t.Run("Sizes are very close (within 10%)", func(t *testing.T) {
		score := ev.evaluateSize(100, 105)
		if score != 15 {
			t.Errorf("expected 15, but got %d", score)
		}
	})

	t.Run("Sizes sre too different", func(t *testing.T) {
		score := ev.evaluateSize(100, 500)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}
