package evaluator

import (
	"encoding/json"
	"math"
	"reflect"
	"sort"
	"strings"
)

type Evaluator struct {
	StatusWeight    int
	StructureWeight int
	KeywordWeight   int
	SizeWeight      int
}

// テスト通していく最中に実装
// 今はテストでモックを書き入れているだけ
type ResponseData struct {
	StatusCode int
	Body       string
}

func (e *Evaluator) evaluateStatus(attackerStatus int, victimStatus int) int {
	if attackerStatus == victimStatus {
		return e.StatusWeight
	}
	return 0
}

func (e *Evaluator) evaluateStructure(victimBody, attackerBody string) int {
	var victimMap, attackerMap map[string]interface{}

	if err := json.Unmarshal([]byte(victimBody), &victimMap); err != nil {
		return 0
	}
	if err := json.Unmarshal([]byte(attackerBody), &attackerMap); err != nil {
		return 0
	}

	vKeys := extractKeys(victimMap)
	aKeys := extractKeys(attackerMap)

	if reflect.DeepEqual(vKeys, aKeys) {
		return e.StructureWeight
	}
	return 0
}

func extractKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (e *Evaluator) evaluateKeywords(body string, keywords []string) int {
	for _, kw := range keywords {
		if strings.Contains(body, kw) {
			return e.KeywordWeight
		}
	}
	return 0
}

// 唯一の公開
func (e *Evaluator) FullEvaluation(victim, attacker ResponseData, keywords []string) int {
	var totalScore int

	totalScore += e.evaluateStatus(attacker.StatusCode, victim.StatusCode)

	totalScore += e.evaluateStructure(victim.Body, attacker.Body)

	totalScore += e.evaluateKeywords(attacker.Body, keywords)

	totalScore += e.evaluateSize(int64(len(victim.Body)), int64(len(attacker.Body)))

	return totalScore
}

func (e *Evaluator) evaluateSize(victimize, attackerSize int64) int {
	if victimize == 0 {
		return 0
	}

	diff := math.Abs(float64(victimize - attackerSize))
	ratio := diff / float64(victimize)

	if ratio <= 0.1 {
		return e.SizeWeight
	}
	return 0
}
