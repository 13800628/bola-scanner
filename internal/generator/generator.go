package generator

type TargetGenerator interface {
	Next() (string, string, bool)
}
