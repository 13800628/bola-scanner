package scanner

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
	"github.com/yuto-isayama/bola-scanner/internal/network"
)

func TestScanner_E2E_BOLA_Detection(t *testing.T) {
	// テスト用脆弱性を持つサーバー
	ts := httptest.NewServer(http.HandlerFunc(mockUserServer))
	defer ts.Close()

	victimData := evaluator.ResponseData{
		Body: `{"id":101, "name":"Victim", "secret":"top-secret"}`,
	}

	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		Client:     http.Client{},
	}

	// evaluator側でこの重みを使ってスコアを計算する(今後はプロダクト側で重みを設定する)
	ev := evaluator.NewEvaluator()
	gen := generator.NewSequentialGenerator(100, 102, "", "")
	testAuthMap := map[string]auth.Authenticator{
		"": &auth.NoAuth{},
	}

	s := NewScanner(ts.URL+"/api/users/{{ID}}", ev, gen, testAuthMap, 2, client)
	s.VictimData = victimData
	s.Keywords = []string{"secret", "Victim"}

	results := s.Run()

	foundBOLA := false
	for _, res := range results {
		fmt.Printf("ID: %s, Score: %f\n", res.ID, res.Score)
		// フラグを立てる条件(今後は拡張の可能性高い)
		if res.ID == ts.URL+"/api/users/101" && res.IsBOLA() {
			foundBOLA = true
		}
	}
	if !foundBOLA {
		t.Error("BOLA was not detected for ID 101, but it should have been.")
	}
}

// テスト用の関数
func mockUserServer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/users/101" {
		fmt.Fprint(w, `{"id": 101, "name": "Victim", "secret": "top-secret"}`)
		return
	}
	fmt.Fprint(w, `{"id": 999, "name": "Other", "secret": "nothing"}`)
}
