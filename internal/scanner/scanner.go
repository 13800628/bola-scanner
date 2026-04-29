package scanner

import (
	"strings"
	"sync"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
)

type Scanner struct {
	Evaluator   *evaluator.Evaluator
	Generator   generator.TargetGenerator
	Auth        auth.Authenticator
	URLTemplate string
	VictimData  evaluator.ResponseData
	WorkerCount int
	// ここに今後HTTPClientなどのついか
}

type ScanResult struct {
	ID    string
	Score int
}

func NewScanner(e *evaluator.Evaluator, g generator.TargetGenerator, a auth.Authenticator, workerCount int) *Scanner {
	return &Scanner{
		Evaluator:   e,
		Generator:   g,
		Auth:        a,
		WorkerCount: workerCount,
	}
}

func (s *Scanner) Run() []ScanResult {
	idChan := make(chan string)
	resultChan := make(chan ScanResult)
	var wg sync.WaitGroup

	// ワーカーの起動
	for i := 0; i < s.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idChan {
				targerURL := s.buildURL(id)

				resultChan <- ScanResult{
					ID:    targerURL,
					Score: 0,
				}
			}
		}()
	}

	go func() {
		for {
			id, hasNext := s.Generator.Next()
			if id != "" {
				idChan <- id
			}
			if !hasNext {
				break
			}
		}
		close(idChan)
	}()

	var results []ScanResult

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		results = append(results, res)
	}
	return results
}

func (s *Scanner) buildURL(id string) string {
	return strings.Replace(s.URLTemplate, "{{ID}}", id, 1)
}
