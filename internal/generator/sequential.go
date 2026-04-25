package generator

import "strconv"

type SequentialGenerator struct {
	currentID int
	endID     int
	prefix    string
}

func NewSequentialGenerator(start, end int, prefix string) *SequentialGenerator {
	return &SequentialGenerator{
		currentID: start,
		endID:     end,
		prefix:    prefix,
	}
}

func (g *SequentialGenerator) Next() (string, bool) {
	if g.currentID > g.endID {
		return "", false
	}

	res := g.prefix + strconv.Itoa(g.currentID)

	g.currentID++

	hasNext := g.currentID <= g.endID

	return res, hasNext
}
