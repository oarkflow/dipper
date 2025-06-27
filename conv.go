package dipper

var defaultDipper = New(Options{Separator: SEPARATOR, Slice: SLICE})

type Fields map[string]any

func Get(obj any, attribute string, groupBy ...string) (any, error) {
	return defaultDipper.Get(obj, attribute, groupBy...)
}

func GetMany(obj any, attributes []string) (Fields, error) {
	return defaultDipper.GetMany(obj, attributes)
}

func Set(obj any, attribute string, new any) error {
	return defaultDipper.Set(obj, attribute, new)
}
