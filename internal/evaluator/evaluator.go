package evaluator

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

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

func (e *Evaluator) evaluateStatus(attackerStatus int, victimStatus int) (int, string) {
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

	if reflect.DeepEqual(vKeys, aKeys) {
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

func (e *Evaluator) evaluateSize(victimSize, attackerSize int64) (int, string) {
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

// 唯一の公開
func (e *Evaluator) FullEvaluation(victim, attacker ResponseData, keywords []string) EvaluationResult {
	res := EvaluationResult{Factors: []string{}, TotalScore: 0}

	// 1. Status Check
	if s, m := e.evaluateStatus(attacker.StatusCode, victim.StatusCode); s > 0 {
		res.TotalScore += s
		res.Factors = append(res.Factors, m)
	}

	// 2. Structure Check
	if s, m := e.evaluateStructure(victim.Body, attacker.Body); s > 0 {
		res.TotalScore += s
		res.Factors = append(res.Factors, m)
	}

	// 3. Keywords Check
	if s, m := e.evaluateKeywords(attacker.Body, keywords); s > 0 {
		res.TotalScore += s
		res.Factors = append(res.Factors, m)
	}

	// 4. Size Check
	if s, m := e.evaluateSize(int64(len(victim.Body)), int64(len(attacker.Body))); s > 0 {
		res.TotalScore += s
		res.Factors = append(res.Factors, m)
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
