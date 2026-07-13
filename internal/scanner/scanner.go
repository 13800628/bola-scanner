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
	AuthMap     map[string]auth.Authenticator
	URLTemplate string
	VictimData  evaluator.ResponseData
	WorkerCount int
	Keywords    []string
	// ここに今後HTTPClientなどのついか
	Client *network.RetryClient
}

type ScanResult struct {
	ID      string
	Score   float64
	Factors []string
}

type ScanJob struct {
	VictimID   string
	AttackerID string
}

func NewScanner(urlTemplate string, e *evaluator.Evaluator, g generator.TargetGenerator, authMap map[string]auth.Authenticator, workerCount int, client *network.RetryClient) *Scanner {
	return &Scanner{
		URLTemplate: urlTemplate,
		Evaluator:   e,
		Generator:   g,
		AuthMap:     authMap,
		WorkerCount: workerCount,
		Client:      client,
	}
}

// スキャン開始前に被害者データを動的に取得するメソッド
func (s *Scanner) PrepareVictimData(victimAuth auth.Authenticator, victimID string) error {
	targetURL := s.buildURL(victimID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return err
	}

	// レスポンスを取得しにいく
	victimAuth.Apply(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	s.VictimData = evaluator.ResponseData{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
	return nil
}

func (s *Scanner) scanOne(job ScanJob) (ScanResult, error) {
	targetURL := s.buildURL(job.VictimID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return ScanResult{}, err
	}

	// Attackerに対応する認証情報があれば適応する
	if authenticator, exists := s.AuthMap[job.AttackerID]; exists {
		authenticator.Apply(req)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return ScanResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScanResult{}, err
	}

	currentData := evaluator.ResponseData{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
	evalRes := s.Evaluator.FullEvaluation(s.VictimData, currentData, s.Keywords)
	return ScanResult{
		ID:      targetURL,
		Score:   float64(evalRes.TotalScore),
		Factors: evalRes.Factors,
	}, nil
}

func (s *Scanner) Run() []ScanResult {
	idChan := make(chan ScanJob)
	resultChan := make(chan ScanResult)
	var wg sync.WaitGroup

	// 深いから修正余地か
	// ワーカーの起動
	for i := 0; i < s.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range idChan {
				res, err := s.scanOne(job)
				if err != nil {
					continue
				}
				resultChan <- res
			}
		}()
	}

	go func() {
		for {
			victimID, attackerID, hasNext := s.Generator.Next()
			if victimID != "" {
				idChan <- ScanJob{VictimID: victimID, AttackerID: attackerID}
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

// 判定する関数だが浅いので今後改善余地か
func (r *ScanResult) IsBOLA() bool {
	return r.Score >= 70
}
