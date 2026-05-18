package generator

import "strconv"

type SequentialGenerator struct {
	currentID  int
	endID      int
	prefix     string
	attackerID string
}

func NewSequentialGenerator(start, end int, prefix string, attackerID string) *SequentialGenerator {
	if start > end {
		panic("start must be less than or equals to end")
	}
	return &SequentialGenerator{
		currentID:  start,
		endID:      end,
		prefix:     prefix,
		attackerID: attackerID,
	}
}

func (g *SequentialGenerator) Next() (string, string, bool) {
	if g.currentID > g.endID {
		return "", "", false
	}

	victim := g.prefix + strconv.Itoa(g.currentID)

	attacker := g.attackerID

	g.currentID++

	return victim, attacker, true
}
