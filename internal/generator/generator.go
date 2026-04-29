package generator

type TargetGenerator interface {
	Next() (string, bool)
}
