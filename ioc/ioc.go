package ioc

var root = CreateContainer()

// Clear calls root Clear method.
func Clear() {
	root.Clear()
}

// MustBind calls root MustBind method.
func MustBind(instance interface{}) {
	root.MustBind(instance)
}

// MustBindWithAlias calls root MustBindWithAlias method.
func MustBindWithAlias(instance interface{}, alias string) {
	root.MustBindWithAlias(instance, alias)
}

// MustBindSingleton calls root MustBindSingleton method.
func MustBindSingleton(resolver interface{}, meta interface{}) {
	root.MustBindSingleton(resolver, meta)
}

// MustBindSingletonWithAlias calls root MustBindSingletonWithAlias method.
func MustBindSingletonWithAlias(resolver interface{}, meta interface{}, alias string) {
	root.MustBindSingletonWithAlias(resolver, meta, alias)
}

// MustBindTransient calls root MustBindTransient method.
func MustBindTransient(resolver interface{}, meta interface{}) {
	root.MustBindTransient(resolver, meta)
}

// MustBindTransientWithAlias calls root MustBindTransientWithAlias method.
func MustBindTransientWithAlias(resolver interface{}, meta interface{}, alias string) {
	root.MustBindTransientWithAlias(resolver, meta, alias)
}

// Resolve calls root Resolve method.
func Resolve(receiver interface{}) error {
	return root.Resolve(receiver)
}

// ResolveWithAlias calls root ResolveWithAlias method.
func ResolveWithAlias(receiver interface{}, alias string) error {
	return root.ResolveWithAlias(receiver, alias)
}
