package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
	"github.com/yuto-isayama/bola-scanner/internal/network"
	"github.com/yuto-isayama/bola-scanner/internal/scanner"
)

func main() {
	fmt.Println("BOLA Scanner Starting...")

	ev := evaluator.NewEvaluator()
	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Client:     http.Client{Timeout: 10 * time.Second},
	}
	urlTmpl := "http://api.example.com/v1/users/{{ID}}"
	keywords := []string{"admin", "password", "email", "sercret", "token"}

	attacker := &auth.Profile{Username: "attacker_user", Password: "password123"}
	victim := &auth.Profile{Username: "victim_user", Password: "password456"}

	baseURL := "http://api.example.com"
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
		if res.Score > 50 {
			foundCount++
			fmt.Printf("[Potential BOLA] Score: %.2f | ID:%s\n", res.Score, res.ID)
			for _, factor := range res.Factors {
				fmt.Printf(" - %s\n", factor)
			}
		}
	}
	fmt.Printf("\n Scan finished in %v. Found %d suspecious endpoints.\n", duration, foundCount)
}

func fetchVictimBaseline(client *network.RetryClient, v *auth.Profile, tmpl string) evaluator.ResponseData {
	url := strings.Replace(tmpl, "{{ID}}", fmt.Sprintf("%d", v.UserID), 1)
	req, _ := http.NewRequest("GET", url, nil)

	v.GetAuthenticator().Apply(req)

	resp, _ := client.Do(req)
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return evaluator.ResponseData{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}
