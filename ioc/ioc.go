package ioc

var root Container = &container{cnt: map[string]binderMap{}}

// CreateContainer ...
func CreateContainer() Container {
	return &container{cnt: map[string]binderMap{}}
}

// TestGetCnt only for testing!
func TestGetCnt() map[string]binderMap {
	return root.(*container).cnt
}

// MustBind ...
func MustBind(instance interface{}) {
	root.MustBind(instance)
}

// MustBindWithAlias ...
func MustBindWithAlias(instance interface{}, alias string) {
	root.MustBindWithAlias(instance, alias)
}

func MustBindSingleton(resolver interface{}, meta interface{}) {
	root.MustBindSingleton(resolver, meta)
}

func MustBindTransient(resolver interface{}, meta interface{}) {
	root.MustBindTransient(resolver, meta)
}

// Resolve ...
func Resolve(receiver interface{}) error {
	return root.Resolve(receiver)
}

// ResolveWithAlias ...
func ResolveWithAlias(receiver interface{}, alias string) error {
	return root.ResolveWithAlias(receiver, alias)
}
