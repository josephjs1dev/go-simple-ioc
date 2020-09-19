package ioc

var root Container = &container{cnt: map[string]map[string]binder{}}

// CreateContainer ...
func CreateContainer() Container {
	return &container{cnt: map[string]map[string]binder{}}
}

// Bind ...
func Bind(instance interface{}) {
	root.Bind(instance)
}

// BindWithAlias ...
func BindWithAlias(instance interface{}, alias string) {
	root.BindWithAlias(instance, alias)
}
