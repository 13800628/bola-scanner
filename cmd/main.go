package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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
	fmt.Println("========================================")
	fmt.Println("         BOLA Scanner Starting...       ")
	fmt.Println("========================================")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter target Base URL (e.g., http://api.example.com)\n[Leave empty to run in Mock Deomo Mode]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var baseURL string
	var mockServer *httptest.Server

	if input != "" {
		baseURL = strings.TrimSuffix(input, "/")
		fmt.Printf("\n[*] Mode: LIVE SCANNING -> %s\n\n", baseURL)
	} else {
		// Mock
		fmt.Println("\n[*] Mode: MOCK DEMO (Using built-in mock server)")
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if r.URL.Path == "/login" {
				w.Write([]byte(`{"token": "mock-token-abcde", "user_id": 999}`))
				return
			}
			w.Write([]byte(`{"id": 100, "name": "tesr_user", "email": "test@example.com"}`))
		}))
		baseURL = mockServer.URL
		defer mockServer.Close()
	}

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

	if err := attacker.Login(baseURL); err != nil {
		log.Fatalf("attacker login failed: %v", err)
	}
	if err := victim.Login(baseURL); err != nil {
		log.Fatalf("victim login failed: %v", err)
	}

	attackerKey := "attacker_attacker"
	gen := generator.NewSequentialGenerator(100, 120, "", attackerKey)

	authMap := map[string]auth.Authenticator{
		attackerKey: attacker.GetAuthenticator(),
		"":          &auth.NoAuth{},
	}

	// スキャナーの作成
	s := scanner.NewScanner(urlTmpl, ev, gen, authMap, 5, client)
	s.Keywords = keywords

	fmt.Println("[*] Fetching victim's baseline data...")
	victimData, err := fetchVictimBaseline(client, victim, urlTmpl)
	if err != nil {
		log.Fatalf("failed to fetch victim baseline: %v", err)
	}
	s.VictimData = victimData

	fmt.Printf("Scanning %s ...\n", urlTmpl)
	startTime := time.Now()
	results := s.Run()
	duration := time.Since(startTime)

	fmt.Println("\n--- Scan Results ---")
	foundCount := 0
	for _, res := range results {
		if res.Score >= 70 {
			foundCount++
			fmt.Printf("[Potential BOLA] Score: %.2f | ID:%s\n", res.Score, res.ID)
		}
	}
	fmt.Printf("\n Scan finished in %v. Found %d suspecious endpoints.\n", duration, foundCount)
}

func fetchVictimBaseline(client *network.RetryClient, v *auth.Profile, tmpl string) (evaluator.ResponseData, error) {
	url := strings.Replace(tmpl, "{{ID}}", fmt.Sprintf("%d", v.UserID), 1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return evaluator.ResponseData{}, err
	}

	v.GetAuthenticator().Apply(req)

	resp, err := client.Do(req)
	if err != nil {
		return evaluator.ResponseData{}, fmt.Errorf("failed to fetch baseline: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return evaluator.ResponseData{}, err
	}

	return evaluator.ResponseData{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}, nil
}
