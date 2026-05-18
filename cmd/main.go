package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
	"github.com/yuto-isayama/bola-scanner/internal/network"
	"github.com/yuto-isayama/bola-scanner/internal/scanner"
)

func main() {
	// サーバーがない状態でも動くようにモックサーバーを使う
	fmt.Println("BOLA Scanner Starting...")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// ログインへのものならダミーのトークンを
		if r.URL.Path == "/login" {
			w.Write([]byte(`{"token": "mock-token-abcde", "user_id": 990}`))
			return
		}

		// それ以外ならダミーユーザーデータを
		w.Write([]byte(`[{"id": 100, "name"" "test_user", "email": "test@example.com"}]`))
	}))
	defer server.Close()

	baseURL := server.URL
	urlTmpl := baseURL + "/v1/users/{{ID}}"

	ev := evaluator.NewEvaluator()
	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		Client:     http.Client{Timeout: 10 * time.Second},
	}

	keywords := []string{"admin", "password", "email", "sercret", "token"}

	attacker := auth.NewProfile("attacker_user", "password123")
	victim := auth.NewProfile("victim_user", "password456")

	fmt.Println("[*] Authenticating users...")
	_ = attacker.Login(baseURL)
	_ = victim.Login(baseURL)

	attackerKey := "attacker_attacker"
	gen := generator.NewSequentialGenerator(100, 120, attackerKey)

	authMap := map[string]auth.Authenticator{
		attackerKey: attacker.GetAuthenticator(),
		"":          &auth.NoAuth{},
	}

	// スキャナーの作成
	s := scanner.NewScanner(urlTmpl, ev, gen, authMap, 5, client)
	s.Keywords = keywords

	fmt.Println("[*] Fetching victim's baseline data...")
	s.VictimData = fetchVictimBaseline(client, victim, urlTmpl)

	fmt.Printf("Scanning %s ...\n", urlTmpl)
	startTime := time.Now()
	results := s.Run()
	duration := time.Since(startTime)

	fmt.Println("\n--- Scan Results ---")
	foundCount := 0
	for _, res := range results {
		if res.Score >= 0 {
			foundCount++
			fmt.Printf("[Potential BOLA] Score: %.2f | ID:%s\n", res.Score, res.ID)
		}
	}
	fmt.Printf("\n Scan finished in %v. Found %d suspecious endpoints.\n", duration, foundCount)
}

func fetchVictimBaseline(client *network.RetryClient, v *auth.Profile, tmpl string) evaluator.ResponseData {
	url := strings.Replace(tmpl, "{{ID}}", fmt.Sprintf("%d", v.UserID), 1)
	req, _ := http.NewRequest("GET", url, nil)

	v.GetAuthenticator().Apply(req)

	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("[-] Failed to fetch baseline data (is the target server running?): %v", err))
	}

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return evaluator.ResponseData{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}
