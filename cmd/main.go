package cmd

import (
	"fmt"
	"net/http"
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
	gen := generator.NewSequentialGenerator(100, 120, "")
	at := &auth.NoAuth{}

	client := &network.RetryClient{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		Client:     http.Client{Timeout: 10 * time.Second},
	}

	urlTmpl := "http://api.example.com/v1/users/{{ID}}"
	keywords := []string{"admin", "password", "email", "sercret", "token"}

	s := scanner.NewScanner(urlTmpl, ev, gen, at, 5, client)
	s.Keywords = keywords

	s.VictimData = evaluator.ResponseData{
		StatusCode: 200,
		Body:       `{"id":99, "name":"test-user", "role":"user"}`,
	}

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
