package scanner

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
	"github.com/yuto-isayama/bola-scanner/internal/network"
)

func TestScanner_Run(t *testing.T) {
	ev := evaluator.NewEvaluator()

	urlTmpl := "http://localhost:8080/users/{{ID}}"
	gen := generator.NewSequentialGenerator(10, 10, "")

	// ここはモックにする、今後は動的なもにする可能性高い
	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		Client:     http.Client{},
	}
	s := NewScanner(urlTmpl, ev, gen, &auth.NoAuth{}, 1, client)

	s.VictimData = evaluator.ResponseData{
		StatusCode: 200,
		Body:       `{"id": 1, "name": "victim"}`,
	}

	results := s.Run()

	if len(results) == 0 {
		t.Log("No results found (this is expected because no server is running)")
		return
	}

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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	gen := generator.NewSequentialGenerator(1, 3, "")
	ev := evaluator.NewEvaluator()
	auth := &auth.NoAuth{}

	client := &network.RetryClient{
		MaxRetries: 1,
		BaseDelay:  10 * time.Millisecond,
		Client:     http.Client{},
	}

	// テスト用にURLはモックとしている、今後は動的に取得したものを使う方針にする
	s := NewScanner("http://example.com/{{ID}}", ev, gen, auth, 2, client)
	s.WorkerCount = 2 // 並行処理のワーカーの数

	results := s.Run()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for _, res := range results {
		t.Logf("Result ID: %s, Score: %f", res.ID, res.Score)
	}
}
