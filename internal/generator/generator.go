package generator

type TaragetGenerator interface {
	Next() (string, bool)
}
