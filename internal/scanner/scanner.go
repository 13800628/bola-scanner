package scanner

import (
	"strings"

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
	// ここに今後HTTPClientなどのついか
}

type ScanResult struct {
	ID    string
	Score int
}

func NewScanner(e *evaluator.Evaluator, g generator.TargetGenerator) *Scanner {
	return &Scanner{
		Evaluator: e,
		Generator: g,
	}
}

func (s *Scanner) Run() []ScanResult {
	var results []ScanResult

	for {
		id, hasNext := s.Generator.Next()
		if id == "" && !hasNext {
			break
		}

		targetURL := s.buildURL(id)

		results = append(results, ScanResult{
			ID:    targetURL,
			Score: 0,
		})

		if !hasNext {
			break
		}
	}

	return results
}

func (s *Scanner) buildURL(id string) string {
	return strings.Replace(s.URLTemplate, "{{ID}}", id, 1)
}
