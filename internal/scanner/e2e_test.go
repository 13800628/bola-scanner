package scanner

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
)

func TestScanner_E2E_BOLA_Detection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users/101" {
			fmt.Fprint(w, `{"id":101, "name":"Victim", "serect":"top-secret"}`)
		} else {
			fmt.Fprint(w, `{"id":unknown, "name":"Other", "secret":"nothing"}`)
		}
	}))
	defer ts.Close()

	victimData := evaluator.ResponseData{
		Body: `{"id":101, "name":"Victim", "secret":"top-secret"}`,
	}

	ev := &evaluator.Evaluator{}
	gen := generator.NewSequentialGenerator(100, 102, "")
	at := &auth.NoAuth{}

	s := NewScanner(ev, gen, at, 2)
	s.URLTemplate = ts.URL + "/api/users/{{ID}}"
	s.VictimData = victimData

	results := s.Run()

	foundBOLA := false
	for _, res := range results {
		// フラグを立てる条件(今後は拡張の可能性高い)
		if res.ID == ts.URL+"/api/users/101" && res.Score > 1.0 {
			foundBOLA = true
		}

		if !foundBOLA {
			t.Error("BOLA was not detected for ID 101, but it should have been.")
		}
	}
}
