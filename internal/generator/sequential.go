package generator

import "strconv"

type SequentialGenerator struct {
	currentID int
	endID     int
	prefix    string
}

func NewSequentialGenerator(start, end int, prefix string) *SequentialGenerator {
	if start > end {
		panic("start must be less than or equals to end")
	}
	return &SequentialGenerator{
		currentID: start,
		endID:     end,
		prefix:    prefix,
	}
}

func (g *SequentialGenerator) Next() (string, string, bool) {
	if g.currentID > g.endID {
		return "", "", false
	}

	victim := g.prefix + strconv.Itoa(g.currentID)

	attacker := ""

	g.currentID++

	hasNext := g.currentID <= g.endID

	return victim, attacker, hasNext
}
