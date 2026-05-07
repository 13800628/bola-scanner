package evaluator

import (
	"fmt"
	"testing"
)

func TestEvaluateStatus(t *testing.T) {
	ev := NewEvaluator()

	t.Run("Status code match gives full score", func(t *testing.T) {
		score, _ := ev.evaluateStatus(200, 200)
		expected := 30

		if score != expected {
			t.Errorf("expected 30, but got %d", score)
		}
	})

	t.Run("Status code mismatch gives zero", func(t *testing.T) {
		score, _ := ev.evaluateStatus(200, 403)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}

// JSON Test
func TestEvaluator_EvaluateStructure(t *testing.T) {
	ev := NewEvaluator()

	t.Run("JSON keys match perfectly", func(t *testing.T) {
		victimJSON := `{"id": 1, "name": "attacker", "email": "v@example.com"}`
		attackerJSON := `{"id": 100, "name": "attacker", "email": "a@example.com"}`

		score, _ := ev.evaluateStructure(victimJSON, attackerJSON)
		if score != 30 {
			t.Errorf("expected 30, but got %d", score)
		}
	})

	t.Run("JSON keys mismatch", func(t *testing.T) {
		victimJSON := `{"id": 1, "name": "victim"}`
		attackerJSON := `{"error": "not found"}` // 全く違う構造

		score, _ := ev.evaluateStructure(victimJSON, attackerJSON)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})

	t.Run("Invalid JSPON returns 0", func(t *testing.T) {
		score, msg := ev.evaluateStructure("invalid", `{"id":1}`)
		if score != 0 {
			t.Errorf("expected 0, got %d", score)
		}
		if msg != "" {
			t.Errorf("expected empty message, got %s", msg)
		}
	})
}

// ここのテストだけは、ステータスコードを403にしてキーワードだけでスコアが上がることを検証する
func TestEvaluator_EvaluateKeyWords(t *testing.T) {
	ev := NewEvaluator()

	t.Run("Response contains sensitive keywords", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{}`}
		attacker := ResponseData{
			StatusCode: 403,
			Body:       `{"id": 100, "name": "Isayama", "note": "secret info"}`,
		}
		keywords := []string{"Isayama", "secret"}

		// ここを実装クラスで実装
		result := ev.FullEvaluation(victim, attacker, keywords)
		if result.TotalScore != 25 {
			t.Errorf("expected 25, but got %d", result.TotalScore)
		}
	})

	t.Run("Response does not contain keywords", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{}`}
		attacker := ResponseData{
			StatusCode: 403,
			Body:       `{"id": 101, "name": "Unknown", "note": "public"}`,
		}
		keywords := []string{"Isayama"}

		result := ev.FullEvaluation(victim, attacker, keywords)
		if result.TotalScore != 0 {
			t.Errorf("expected 0, but got %d", result.TotalScore)
		}
	})
}

func TestEvaluator_FullEvaluation(t *testing.T) {
	ev := NewEvaluator()

	t.Run("Perfect Match - 100 points", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		keywords := []string{"victim"}

		result := ev.FullEvaluation(victim, attacker, keywords)

		if result.TotalScore != 100 {
			t.Errorf("expected 100, but got %d", result.TotalScore)
		}
	})

	t.Run("High probably BOLA (All criteria met)", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim"}`}
		keywords := []string{"victim"}

		result := ev.FullEvaluation(victim, attacker, keywords)

		if result.TotalScore != 100 {
			t.Errorf("expected 100, but got %d", result.TotalScore)
		}
	})

	t.Run("Low probably (Only status matches)", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: `{"id": 1, "name": "victim", "email": "v@example.com"}`}
		attacker := ResponseData{StatusCode: 200, Body: `{"status": "error", "message": "unauthorized access detected"}`}
		keywords := []string{"victim"}

		result := ev.FullEvaluation(victim, attacker, keywords)

		if result.TotalScore != 30 {
			t.Errorf("expected 30, but got %d", result.TotalScore)
		}
	})

	t.Run("Non-JSON body only status matches", func(t *testing.T) {
		victim := ResponseData{StatusCode: 200, Body: "plain text"}
		attacker := ResponseData{StatusCode: 200, Body: "plain text"}
		keywords := []string{}

		result := ev.FullEvaluation(victim, attacker, keywords)

		if result.TotalScore != 45 {
			t.Errorf("expected 45, got %d", result.TotalScore)
		}
	})
}

func TestEvaluator_FullEvaluation_Reasoning(t *testing.T) {
	ev := NewEvaluator()

	victim := ResponseData{StatusCode: 200, Body: "..."}
	attacker := ResponseData{StatusCode: 200, Body: "secret data found"}
	keywords := []string{"secret"}

	result := ev.FullEvaluation(victim, attacker, keywords)

	if result.TotalScore != 55 {
		t.Errorf("expected score 55, got %d", result.TotalScore)
	}

	fmt.Println("Detection Reasons:")
	for _, f := range result.Factors {
		fmt.Printf("- %s\n", f)
	}

	if len(result.Factors) == 0 {
		t.Errorf("expected 2 factors, got %d", len(result.Factors))
	}
}

func TestEvaluator_EvaluateSize(t *testing.T) {
	ev := NewEvaluator()

	t.Run("Sizes are very close (within 10%)", func(t *testing.T) {
		score, _ := ev.evaluateSize(100, 105)
		if score != 15 {
			t.Errorf("expected 15, but got %d", score)
		}
	})

	t.Run("Sizes sre too different", func(t *testing.T) {
		score, _ := ev.evaluateSize(100, 500)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})

	t.Run("Victim size is zero returns 0", func(t *testing.T) {
		score, _ := ev.evaluateSize(0, 100)
		if score != 0 {
			t.Errorf("expected 0, but got %d", score)
		}
	})
}
