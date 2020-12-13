package ioc

var root = CreateContainer()

// Clear calls root Clear method.
func Clear() {
	root.Clear()
}

// MustBindSingleton calls root MustBindSingleton method.
func MustBindSingleton(resolver interface{}, opts ...BindOption) {
	root.MustBindSingleton(resolver, opts...)
}

// MustBindTransient calls root MustBindTransient method.
func MustBindTransient(resolver interface{}, opts ...BindOption) {
	root.MustBindTransient(resolver, opts...)
}

// Resolve calls root Resolve method.
func Resolve(receiver interface{}, opts ...ResolveOption) error {
	return root.Resolve(receiver, opts...)
}

// MustResolve calls root MustResolve method.
func MustResolve(receiver interface{}, opts ...ResolveOption) {
	root.MustResolve(receiver, opts...)
}
