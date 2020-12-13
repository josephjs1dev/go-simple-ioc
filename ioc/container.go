package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const structTag = "ioc"
const defaultAlias = "default"

var (
	ErrNotRegistered             = errors.New("information is not registered to container")
	ErrAliasNotKnown             = errors.New("alias is not known")
	ErrInstanceMustNotBeFunction = errors.New("instance must not be a function")
)

// Container provides utility functions to bind and resolve.
type Container interface {
	Clear()
	BindSingleton(interface{}, ...BindOption) error
	MustBindSingleton(interface{}, ...BindOption)
	BindTransient(interface{}, ...BindOption) error
	MustBindTransient(interface{}, ...BindOption)
	Resolve(interface{}, ...ResolveOption) error
	MustResolve(interface{}, ...ResolveOption)
}

type binder struct {
	// isSingleton is flag to check whether it is singleton or transient.
	isSingleton bool
	// resolver is internal function that resolves the actual implementation.
	resolver interface{}
	// meta is metadata information of the instance.
	meta interface{}
	// instance is actual implementation saved.
	instance interface{}
	// dependencies is a list of dependency from the implementation.
	dependencies [][2]string
}

type binderMap map[string]*binder

// Implementation of Container interface.
type container struct {
	// Map of string to map of string interface.
	// First key is the type (can be interface or struct) while second key is alias (default is default key)
	// to the implementation.
	cnt map[string]binderMap
}

// CreateContainer creates new struct that implements Container interface.
func CreateContainer() Container {
	return &container{cnt: map[string]binderMap{}}
}

func getLabel(p reflect.Type) string {
	return p.String()
}

func resolveTypePtrNonFunc(instance interface{}) (reflect.Type, error) {
	instanceType := reflect.TypeOf(instance)
	if instanceType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("expected pointer, but instead got %v", instanceType)
	}

	instanceType = instanceType.Elem()
	// Return error as we don't allow function type when binding.
	if instanceType.Kind() == reflect.Func {
		return nil, ErrInstanceMustNotBeFunction
	}

	return instanceType, nil
}

// Clear clears root / default container internal data.
// Does not handles multiple threads.
func (c *container) Clear() {
	c.cnt = map[string]binderMap{}
}

type bindOption struct {
	alias       string
	meta        interface{}
	isSingleton bool
}

type BindOption func(o *bindOption)

func WithBindAlias(alias string) BindOption {
	return func(opt *bindOption) {
		opt.alias = alias
	}
}

func WithBindMeta(meta interface{}) BindOption {
	return func(opt *bindOption) {
		opt.meta = meta
	}
}

type resolveOption struct {
	alias string
}

type ResolveOption func(o *resolveOption)

func WithResolveAlias(alias string) ResolveOption {
	return func(o *resolveOption) {
		o.alias = alias
	}
}

func applyBindOption(o *bindOption, opts []BindOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func getDependencies(resolverType reflect.Type, instanceType reflect.Type) [][2]string {
	dependencyLabelMap := map[string]int{}
	for idx := 0; idx < resolverType.NumIn(); idx++ {
		paramType := resolverType.In(idx)
		dependencyLabelMap[getLabel(paramType)] = idx
	}

	dependencies := make([][2]string, resolverType.NumIn())
	for idx := 0; idx < instanceType.NumField(); idx++ {
		field := instanceType.Field(idx)
		label := getLabel(field.Type)
		inIdx, ok := dependencyLabelMap[label]
		if !ok {
			continue
		}

		tag, ok := field.Tag.Lookup(structTag)
		if !ok || tag == "" {
			tag = defaultAlias
		}
		delete(dependencyLabelMap, label)

		v := strings.Split(tag, ",")
		dependencies[inIdx] = [2]string{label, v[0]}
	}

	// Leftover will be set to default
	for label, inIdx := range dependencyLabelMap {
		dependencies[inIdx] = [2]string{label, defaultAlias}
	}

	return dependencies
}

func (c *container) bind(resolver interface{}, opt *bindOption) error {
	resolverType := reflect.TypeOf(resolver)
	// Must be a function.
	if resolverType.Kind() != reflect.Func {
		return fmt.Errorf("expected resolver to be function, but instead got %v", resolverType.Kind())
	}
	if resolverType.NumOut() < 1 {
		return fmt.Errorf("expected minimum output of 1, but instead got: %v", resolverType.NumOut())
	}

	instanceType := resolverType.Out(0)
	if instanceType.Kind() != reflect.Ptr && instanceType.Kind() != reflect.Interface {
		return fmt.Errorf("expected pointer or interface, but instead got %v", instanceType)
	}

	label := getLabel(instanceType)

	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}
	if opt.meta != nil && instanceType.Kind() == reflect.Interface {
		metaType := reflect.TypeOf(opt.meta)
		if metaType.Kind() != reflect.Ptr {
			return fmt.Errorf("expected meta to be pointer, but instead got %v", metaType.Kind())
		}
		if !metaType.Implements(instanceType) {
			return fmt.Errorf("%v does not implement %v", metaType.Kind(), instanceType.Kind())
		}

		instanceType = metaType.Elem()
	}

	dependencies := getDependencies(resolverType, instanceType)
	if v, ok := c.cnt[label]; !ok {
		c.cnt[label] = binderMap{
			opt.alias: {isSingleton: opt.isSingleton, resolver: resolver, meta: opt.meta, dependencies: dependencies},
		}
	} else {
		v[opt.alias] = &binder{isSingleton: opt.isSingleton, resolver: resolver, meta: opt.meta, dependencies: dependencies}
	}

	return nil
}

// BindSingleton binds given resolver function and metadata information to container with singleton flag.
// As it is singleton, after first resolve, container will save resolved information and immediately returns data
// for next resolve.
// Resolver must be a function that returns interface or pointer struct and meta can be nil or must implements
// returned interface type from resolver.
func (c *container) BindSingleton(resolver interface{}, opts ...BindOption) error {
	o := &bindOption{alias: defaultAlias, isSingleton: true}
	applyBindOption(o, opts)

	return c.bind(resolver, o)
}

// MustBindSingleton is same as BindSingleton, but will panic if error.
func (c *container) MustBindSingleton(resolver interface{}, opts ...BindOption) {
	if err := c.BindSingleton(resolver, opts...); err != nil {
		panic(err)
	}
}

// BindTransient binds given resolver function and metadata information to container without singleton flag.
// Each resolve will create new object.
// Resolver must be a function that returns interface or pointer struct and meta can be nil or must implements
// returned interface type from resolver.
func (c *container) BindTransient(resolver interface{}, opts ...BindOption) error {
	o := &bindOption{alias: defaultAlias, isSingleton: false}
	applyBindOption(o, opts)

	return c.bind(resolver, o)
}

// MustBindTransient is same as BindTransient, but will panic if error.
func (c *container) MustBindTransient(resolver interface{}, opts ...BindOption) {
	if err := c.BindTransient(resolver, opts...); err != nil {
		panic(err)
	}
}

func (c *container) getBinder(label, binderLabel string) (*binder, error) {
	binderMap, ok := c.cnt[label]
	if !ok {
		return nil, ErrNotRegistered
	}

	binder, ok := binderMap[binderLabel]
	if !ok {
		return nil, ErrAliasNotKnown
	}

	return binder, nil
}

func applyResolveOption(o *resolveOption, opts []ResolveOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (c *container) resolve(receiver interface{}, label string, opt *resolveOption) (err error) {
	receiverType, err := resolveTypePtrNonFunc(receiver)
	if err != nil {
		return err
	}

	if label == "" {
		label = getLabel(receiverType)
	}
	binder, err := c.getBinder(label, opt.alias)
	if err != nil {
		return err
	}

	receiverValue := reflect.ValueOf(receiver).Elem()
	if binder.instance != nil {
		receiverValue.Set(reflect.ValueOf(binder.instance).Elem())

		return nil
	}

	resolverType := reflect.TypeOf(binder.resolver)
	in := make([]reflect.Value, 0)
	for idx := 0; idx < resolverType.NumIn(); idx++ {
		paramType := resolverType.In(idx)
		paramValue := reflect.New(paramType).Interface()

		dependency := binder.dependencies[idx]
		if err := c.resolve(&paramValue, dependency[0], &resolveOption{alias: dependency[1]}); err != nil {
			return fmt.Errorf("failed to resolve inner label %v with alias %v: %w", dependency[0], dependency[1], err)
		}

		in = append(in, reflect.ValueOf(paramValue))
	}

	resolverValue := reflect.ValueOf(binder.resolver)
	receiverValue.Set(resolverValue.Call(in)[0])

	if binder.isSingleton {
		binder.instance = reflect.ValueOf(receiver).Interface()
	}

	return nil
}

// Resolve resolves given receiver to appropriate bound information in container.
// Will returns ErrNotRegistered, ErrAliasNotKnown, or any relevant errors if failed to resolve.
func (c *container) Resolve(receiver interface{}, opts ...ResolveOption) (err error) {
	o := &resolveOption{alias: defaultAlias}
	applyResolveOption(o, opts)

	return c.resolve(receiver, "", o)
}

// Resolve resolves given receiver to appropriate bound information in container.
// Will panic if failed to resolve.
func (c *container) MustResolve(receiver interface{}, opts ...ResolveOption) {
	if err := c.Resolve(receiver, opts...); err != nil {
		panic(err)
	}
}
