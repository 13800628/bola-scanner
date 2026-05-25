package generator

type PairGenerator struct {
	ids []string // 全IDリスト
	i   int
	j   int
}

func NewPairGenerator(ids []string) *PairGenerator {
	return &PairGenerator{
		ids: ids,
		i:   0,
		j:   1,
	}
}

func (g *PairGenerator) Next() (victim, attacker string, hasNext bool) {
	// IDが一つしかない部分でのガード節
	if len(g.ids) < 2 {
		return "", "", false
	}

	// iが最後まで来てしまった(ペアがいない)のガード節
	if g.i >= len(g.ids)-1 {
		return "", "", false
	}

	victim = g.ids[g.i]
	attacker = g.ids[g.j]

	g.j++

	if g.j >= len(g.ids) {
		g.i++
		g.j = g.i + 1
	}

	hasNext = g.i < len(g.ids)-1

	return victim, attacker, hasNext
}
