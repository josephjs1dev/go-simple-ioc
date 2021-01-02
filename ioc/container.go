package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const structTagKey = "ioc"
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
	// resolveFunc is internal function that resolves the actual implementation.
	resolveFunc interface{}
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

func getDependencies(resolveFuncType reflect.Type, instanceType reflect.Type) [][2]string {
	labelMap := map[string][]int{}
	labelCtrMap := make(map[string]int)
	for idx := 0; idx < resolveFuncType.NumIn(); idx++ {
		paramType := resolveFuncType.In(idx)
		label := getLabel(paramType)
		if _, ok := labelMap[label]; !ok {
			labelMap[label] = []int{idx}
			labelCtrMap[label] = 0
		} else {
			labelMap[label] = append(labelMap[label], idx)
		}
	}

	dependencies := make([][2]string, resolveFuncType.NumIn())
	if instanceType.Kind() != reflect.Interface {
		for idx := 0; idx < instanceType.NumField(); idx++ {
			field := instanceType.Field(idx)
			label := getLabel(field.Type)
			inIdxList, ok := labelMap[label]
			if !ok {
				continue
			}
			inIdx := inIdxList[labelCtrMap[label]]
			labelCtrMap[label]++

			tag, ok := field.Tag.Lookup(structTagKey)
			v := strings.Split(tag, ",")

			alias := v[0]
			if alias == "" {
				alias = defaultAlias
			}

			dependencies[inIdx] = [2]string{label, alias}
		}
	}

	// Leftover will be set to default
	for label, inIdxList := range labelMap {
		for i := labelCtrMap[label]; i < len(inIdxList); i++ {
			dependencies[inIdxList[i]] = [2]string{label, defaultAlias}
		}
	}

	return dependencies
}

func (c *container) bind(resolveFunc interface{}, opt *bindOption) error {
	resolveFuncType := reflect.TypeOf(resolveFunc)
	// Must be a function.
	if resolveFuncType.Kind() != reflect.Func {
		return fmt.Errorf("expected first params to be function, but instead got %v", resolveFuncType.Kind())
	}
	if resolveFuncType.NumOut() < 1 {
		return fmt.Errorf("expected minimum output of 1, but instead got: %v", resolveFuncType.NumOut())
	}

	instanceType := resolveFuncType.Out(0)
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

	dependencies := getDependencies(resolveFuncType, instanceType)
	if v, ok := c.cnt[label]; !ok {
		c.cnt[label] = binderMap{
			opt.alias: {isSingleton: opt.isSingleton, resolveFunc: resolveFunc, meta: opt.meta, dependencies: dependencies},
		}
	} else {
		v[opt.alias] = &binder{isSingleton: opt.isSingleton, resolveFunc: resolveFunc, meta: opt.meta, dependencies: dependencies}
	}

	return nil
}

// BindSingleton binds given resolve function and metadata information to container with singleton flag.
// As it is singleton, after first resolve, container will save resolved information and immediately returns data
// for next resolve.
// First parameter must be a function that returns interface or pointer struct and meta can be nil or must implements
// returned interface type from resolveFunc.
func (c *container) BindSingleton(resolveFunc interface{}, opts ...BindOption) error {
	o := &bindOption{alias: defaultAlias, isSingleton: true}
	applyBindOption(o, opts)

	return c.bind(resolveFunc, o)
}

// MustBindSingleton is same as BindSingleton, but will panic if error.
func (c *container) MustBindSingleton(resolveFunc interface{}, opts ...BindOption) {
	if err := c.BindSingleton(resolveFunc, opts...); err != nil {
		panic(err)
	}
}

// BindTransient binds given resolveFunc function and metadata information to container without singleton flag.
// Each resolve will create new object.
// First parameter must be a function that returns interface or pointer struct and meta can be nil or must implements
// returned interface type from resolveFunc.
func (c *container) BindTransient(resolveFunc interface{}, opts ...BindOption) error {
	o := &bindOption{alias: defaultAlias, isSingleton: false}
	applyBindOption(o, opts)

	return c.bind(resolveFunc, o)
}

// MustBindTransient is same as BindTransient, but will panic if error.
func (c *container) MustBindTransient(resolveFunc interface{}, opts ...BindOption) {
	if err := c.BindTransient(resolveFunc, opts...); err != nil {
		panic(err)
	}
}

func (c *container) getBinder(label, binderLabel string) (*binder, error) {
	binderMap, ok := c.cnt[label]
	if !ok {
		return nil, fmt.Errorf("can't find dependencies from label %v, err: %w", label, ErrNotRegistered)
	}

	binder, ok := binderMap[binderLabel]
	if !ok {
		return nil, fmt.Errorf("can't find dependencies from label %v with alias %v, err: %w",
			label, binderLabel, ErrAliasNotKnown)
	}

	return binder, nil
}

func applyResolveOption(o *resolveOption, opts []ResolveOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (c *container) buildDependencyArguments(b *binder) ([]reflect.Value, error) {
	resolveFuncType := reflect.TypeOf(b.resolveFunc)
	in := make([]reflect.Value, 0)
	for idx := 0; idx < resolveFuncType.NumIn(); idx++ {
		dependency := b.dependencies[idx]
		argBinder, err := c.getBinder(dependency[0], dependency[1])
		if err != nil {
			return nil, err
		}

		res, err := c.invoke(argBinder)
		if err != nil {
			return nil, err
		}

		in = append(in, reflect.ValueOf(res))
	}

	return in, nil
}

func (c *container) invoke(b *binder) (interface{}, error) {
	if b.instance != nil {
		return b.instance, nil
	}

	args, err := c.buildDependencyArguments(b)
	if err != nil {
		return nil, err
	}

	resolveFuncValue := reflect.ValueOf(b.resolveFunc)
	results := resolveFuncValue.Call(args)

	if b.isSingleton {
		b.instance = results[0].Interface()
	}

	return results[0].Interface(), nil
}

func (c *container) resolve(receiver interface{}, label string, opt *resolveOption) (err error) {
	receiverType, err := resolveTypePtrNonFunc(receiver)
	if err != nil {
		return err
	}

	if label == "" {
		label = getLabel(receiverType)
	}
	b, err := c.getBinder(label, opt.alias)
	if err != nil {
		return err
	}

	receiverValue := reflect.ValueOf(receiver).Elem()
	result, err := c.invoke(b)
	receiverValue.Set(reflect.ValueOf(result))

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
