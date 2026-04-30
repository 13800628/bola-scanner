package scanner

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
	"github.com/yuto-isayama/bola-scanner/internal/evaluator"
	"github.com/yuto-isayama/bola-scanner/internal/generator"
	"github.com/yuto-isayama/bola-scanner/internal/network"
)

type Scanner struct {
	Evaluator   *evaluator.Evaluator
	Generator   generator.TargetGenerator
	Auth        auth.Authenticator
	URLTemplate string
	VictimData  evaluator.ResponseData
	WorkerCount int
	// ここに今後HTTPClientなどのついか
	Client *network.RetryClient
}

type ScanResult struct {
	ID    string
	Score float64
}

func NewScanner(e *evaluator.Evaluator, g generator.TargetGenerator, a auth.Authenticator, workerCount int, client *network.RetryClient) *Scanner {
	return &Scanner{
		Evaluator:   e,
		Generator:   g,
		Auth:        a,
		WorkerCount: workerCount,
		Client:      client,
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

				// リクエストの作成
				req, _ := http.NewRequest("GET", targerURL, nil)

				// 認証情報の適応
				s.Auth.Apply(req)

				resp, err := s.Client.Do(req)
				if err != nil {
					continue // 通信エラー時はスキップ
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				currentData := evaluator.ResponseData{
					Body:       string(body),
					StatusCode: resp.StatusCode,
				}
				score := float64(s.Evaluator.FullEvaluation(s.VictimData, currentData, nil))

				resultChan <- ScanResult{
					ID:    targerURL,
					Score: score,
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
