package generator

type PairGenerator struct {
	ids []string
	i   int
	j   int
}

func NewPairGenerator(ids []string) *PairGenerator
func (g *PairGenerator) Next() (victim, attacker string, hasNext bool)
