package ioc

var root = CreateContainer()

// Clear clears root / default container internal data.
func Clear() {
	root.Clear()
}

// MustBind binds given instance to container and will panic if failed to bind.
func MustBind(instance interface{}) {
	root.MustBind(instance)
}

// MustBindWithAlias binds given instance to container with alias and will panic if failed to bind.
func MustBindWithAlias(instance interface{}, alias string) {
	root.MustBindWithAlias(instance, alias)
}

// MustBindSingleton binds given resolver function and metadata information to container with singleton flag.
// As it is singleton, after first resolve, container will save resolved information and immediately returns data
// for next resolve.
func MustBindSingleton(resolver interface{}, meta interface{}) {
	root.MustBindSingleton(resolver, meta)
}

// MustBindTransient binds given resolver function and metadata information to container without singleton flag.
// Each resolve will create new object.
func MustBindTransient(resolver interface{}, meta interface{}) {
	root.MustBindTransient(resolver, meta)
}

// Resolve resolves given receiver to appropriate bound data in container.
func Resolve(receiver interface{}) error {
	return root.Resolve(receiver)
}

// ResolveWithAlias resolves given receiver and alias to appropriate bound data in container.
func ResolveWithAlias(receiver interface{}, alias string) error {
	return root.ResolveWithAlias(receiver, alias)
}
