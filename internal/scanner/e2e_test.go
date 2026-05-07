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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// ここの条件分岐は今後拡張の可能性高いし、条件が重くなる可能性もあるため、別関数として切り出すことも検討
		if r.URL.Path == "/api/users/101" {
			fmt.Fprint(w, `{"id":101, "name":"Victim", "secret":"top-secret"}`)
		} else {
			fmt.Fprint(w, `{"id":unknown, "name":"Other", "secret":"nothing"}`)
		}
	}))
	defer ts.Close()

	victimData := evaluator.ResponseData{
		Body: `{"id":101, "name":"Victim", "secret":"top-secret"}`,
	}

	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		Client:     http.Client{},
	}

	// 重みの設定しないとスコアが０になる(ここを忘れていたためにBOLAが検出できていなかった)
	// evaluator側でこの重みを使ってスコアを計算する(今後はプロダクト側で重みを設定する)
	ev := &evaluator.Evaluator{
		StatusWeight:    10,
		StructureWeight: 10,
		KeywordWeight:   10,
		SizeWeight:      10,
	}
	gen := generator.NewSequentialGenerator(100, 102, "")
	// テストが通らない
	at := &auth.NoAuth{}

	s := NewScanner(ts.URL+"/api/users/{{ID}}", ev, gen, at, 2, client)
	s.VictimData = victimData

	results := s.Run()

	foundBOLA := false
	for _, res := range results {
		fmt.Printf("ID: %s, Score: %f\n", res.ID, res.Score)
		// フラグを立てる条件(今後は拡張の可能性高い)
		if res.ID == ts.URL+"/api/users/101" && res.Score > 1.0 {
			foundBOLA = true
		}
	}
	if !foundBOLA {
		t.Error("BOLA was not detected for ID 101, but it should have been.")
	}
}
