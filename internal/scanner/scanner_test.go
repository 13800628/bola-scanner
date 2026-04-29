package scanner

import (
	"testing"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
)

func TestScanner_Run(t *testing.T) {
	ev := &evaluator.Evaluator{}

	urlTmpl := "http://localhost:8080/users/{{ID}}"
	gen := generator.NewSequentialGenerator(10, 10, "")
	s := NewScanner(ev, gen, &auth.NoAuth{}, 1)
	s.URLTemplate = urlTmpl

	results := s.Run()

	expectedURL := "http://localhost:8080/users/10"
	if results[0].ID != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, results[0].ID)
	}
}

func TestScanner_BuildURL(t *testing.T) {
	s := &Scanner{
		// サイトによって変化するためテスト段階ではここを変更
		// のちに動的取得にするため変更か？
		URLTemplate: "http://localhost:8080/api/v1/users/{{ID}}/profile",
	}

	expectedResult := "http://localhost:8080/api/v1/users/1001/profile"

	actual := s.buildURL("1001")

	if actual != expectedResult {
		t.Errorf("expected %s, got %s", expectedResult, actual)
	}
}

// 並行処理のテスト
func TestScanner_RunParallel(t *testing.T) {
	gen := generator.NewSequentialGenerator(1, 3, "")
	ev := &evaluator.Evaluator{}
	auth := &auth.NoAuth{}

	s := NewScanner(ev, gen, auth, 2)
	s.URLTemplate = "http://example.com/{{ID}}"
	s.WorkerCount = 2 // 並行処理のワーカーの数

	results := s.Run()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
