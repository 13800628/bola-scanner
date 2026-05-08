package evaluator

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

// 準備スペース
type EvaluationResult struct {
	TotalScore int
	Factors    []string
}
type Evaluator struct {
	StatusWeight    int
	StructureWeight int
	KeywordWeight   int
	SizeWeight      int
}

// 重みづけをハードコート(外部からの実装も将来的には可能にするかも)
func NewEvaluator() *Evaluator {
	return &Evaluator{
		StatusWeight:    30,
		StructureWeight: 30,
		KeywordWeight:   25,
		SizeWeight:      15,
	}
}

// テスト通していく最中に実装
// 今はテストでモックを書き入れているだけ
type ResponseData struct {
	StatusCode int
	Body       string
}

// 実装ブロック
func (e *Evaluator) evaluateStatus(victimStatus int, attackerStatus int) (int, string) {
	if attackerStatus == victimStatus {
		return e.StatusWeight, fmt.Sprintf("Status Code Matched: %d", attackerStatus)
	}
	return 0, ""
}

func (e *Evaluator) evaluateStructure(victimBody, attackerBody string) (int, string) {
	var victimMap, attackerMap map[string]interface{}

	if err := json.Unmarshal([]byte(victimBody), &victimMap); err != nil {
		return 0, ""
	}
	if err := json.Unmarshal([]byte(attackerBody), &attackerMap); err != nil {
		return 0, ""
	}

	vKeys := extractKeys(victimMap)
	aKeys := extractKeys(attackerMap)

	if slices.Equal(vKeys, aKeys) {
		return e.StructureWeight, fmt.Sprintf("JSON Structure Matched: keys=%v", vKeys)
	}
	return 0, ""
}

func (e *Evaluator) evaluateKeywords(body string, keywords []string) (int, string) {
	for _, kw := range keywords {
		if strings.Contains(body, kw) {
			return e.KeywordWeight, fmt.Sprintf("Keyword Found: %s", kw)
		}
	}
	return 0, ""
}

func (e *Evaluator) evaluateSize(victimSize, attackerSize int) (int, string) {
	if victimSize == 0 {
		return 0, ""
	}

	diff := math.Abs(float64(victimSize - attackerSize))
	ratio := diff / float64(victimSize)

	if ratio <= 0.1 {
		return e.SizeWeight, fmt.Sprintf("Body Size Similarity: %.2f%% diff (Victim:%d, Attacker:%d)", ratio*100, victimSize, attackerSize)
	}
	return 0, ""
}

type evaluatorFunc func() (int, string)

// 唯一の公開
func (e *Evaluator) FullEvaluation(victim, attacker ResponseData, keywords []string) EvaluationResult {
	res := EvaluationResult{Factors: []string{}, TotalScore: 0}

	evaluators := []evaluatorFunc{
		func() (int, string) { return e.evaluateStatus(victim.StatusCode, attacker.StatusCode) },
		func() (int, string) { return e.evaluateStructure(victim.Body, attacker.Body) },
		func() (int, string) { return e.evaluateKeywords(attacker.Body, keywords) },
		func() (int, string) { return e.evaluateSize(len(victim.Body), len(attacker.Body)) },
	}

	for _, fn := range evaluators {
		if s, m := fn(); s > 0 {
			res.TotalScore += s
			res.Factors = append(res.Factors, m)
		}
	}

	return res
}

func extractKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
