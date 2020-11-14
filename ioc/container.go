package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const structTag = "ioc"

var (
	ErrNotRegistered             = errors.New("information is not registered to container")
	ErrAliasNotKnown             = errors.New("alias is not known")
	ErrInstanceMustNotBeFunction = errors.New("instance must not be a function")
)

// Container ...
type Container interface {
	MustBind(instance interface{})
	MustBindWithAlias(instance interface{}, alias string)
	Resolve(instance interface{}) (err error)
	ResolveWithAlias(instance interface{}, alias string) (err error)
	MustBindSingleton(resolver interface{}, meta interface{})
	MustBindTransient(resolver interface{}, meta interface{})
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

type binderMap map[string]binder

// Implementation of Container interface.
type container struct {
	// Map of string to map of string interface.
	// First key is the type (can be interface or struct) while second key is alias (default is default key)
	// to the implementation.
	cnt map[string]binderMap
}

func getLabel(p reflect.Type) string {
	return p.String()
}

func getBinderLabel(override string) string {
	if override != "" {
		return override
	}

	return "default"
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

func (c *container) bind(instance interface{}, alias string) error {
	instanceType, err := resolveTypePtrNonFunc(instance)
	if err != nil {
		return err
	}

	label := getLabel(instanceType)
	binderLabel := getBinderLabel(alias)

	if v, ok := c.cnt[label]; !ok {
		c.cnt[label] = map[string]binder{binderLabel: {isSingleton: true, instance: instance}}
	} else {
		v[binderLabel] = binder{isSingleton: true, instance: instance}
	}

	return nil
}

// MustBind does the binding of actual object (struct, not interface) without any alias.
func (c *container) MustBind(instance interface{}) {
	if err := c.bind(instance, ""); err != nil {
		panic(err)
	}
}

// MustBindWithAlias does the binding of actual object (struct, not interface) but with an alias.
func (c *container) MustBindWithAlias(instance interface{}, alias string) {
	if err := c.bind(instance, alias); err != nil {
		panic(err)
	}
}

func getDependencies(resolverType reflect.Type, instanceType reflect.Type) [][2]string {
	dependencyLabelMap := map[string]int{}
	for idx := 0; idx < resolverType.NumIn(); idx++ {
		paramType := resolverType.In(0)
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
		if !ok || tag != "" {
			tag = "default"
		}

		v := strings.Split(tag, ",")
		dependencies[inIdx] = [2]string{label, v[0]}
	}

	return dependencies
}

func (c *container) bindFunc(resolver interface{}, meta interface{}, isSingleton bool, alias string) error {
	resolverType := reflect.TypeOf(resolver)
	// Must be a function.
	if resolverType.Kind() != reflect.Func {
		return fmt.Errorf("expected resolver to be function, but instead got %v", resolverType.Kind())
	}
	if resolverType.NumOut() < 1 {
		return fmt.Errorf("expected minimum output of 1, but instead got: %v", resolverType.NumOut())
	}

	instanceType := resolverType.Out(0)
	label := getLabel(instanceType)
	if meta != nil && instanceType.Kind() == reflect.Interface {
		metaType := reflect.TypeOf(meta)
		if metaType.Kind() != reflect.Ptr {
			return fmt.Errorf("expected meta to be pointer, but instead got %v", metaType.Kind())
		}
		if !metaType.Implements(instanceType) {
			return fmt.Errorf("%v does not implement %v", metaType.Kind(), instanceType.Kind())
		}

		instanceType = metaType.Elem()
	}

	dependencies := getDependencies(resolverType, instanceType)
	binderLabel := getBinderLabel(alias)

	if v, ok := c.cnt[label]; !ok {
		c.cnt[label] = binderMap{
			binderLabel: {isSingleton: isSingleton, resolver: resolver, meta: meta, dependencies: dependencies},
		}
	} else {
		v[binderLabel] = binder{isSingleton: isSingleton, resolver: resolver, meta: meta, dependencies: dependencies}
	}

	return nil
}

func (c *container) mustBindFunc(resolver interface{}, meta interface{}, isSingleton bool, alias string) {
	if err := c.bindFunc(resolver, meta, isSingleton, alias); err != nil {
		panic(err)
	}
}

func (c *container) MustBindSingleton(resolver interface{}, meta interface{}) {
	c.mustBindFunc(resolver, meta, true, "")
}

func (c *container) MustBindSingletonWithAlias(resolver interface{}, meta interface{}, alias string) {
	c.mustBindFunc(resolver, meta, true, alias)
}

func (c *container) MustBindTransient(resolver interface{}, meta interface{}) {
	c.mustBindFunc(resolver, meta, false, "")
}

func (c *container) MustBindTransientWithAlias(resolver interface{}, meta interface{}, alias string) {
	c.mustBindFunc(resolver, meta, false, alias)
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

	return &binder, nil
}

func (c *container) resolve(receiver interface{}, alias string) (err error) {
	receiverType, err := resolveTypePtrNonFunc(receiver)
	if err != nil {
		return err
	}

	label := getLabel(receiverType)
	binderLabel := getBinderLabel(alias)

	binder, err := c.getBinder(label, binderLabel)
	if err != nil {
		return err
	}

	if binder.instance != nil {
		receiverValue := reflect.ValueOf(receiver).Elem()
		receiverValue.Set(reflect.ValueOf(binder.instance).Elem())

		return nil
	}

	return nil
}

func (c *container) Resolve(receiver interface{}) (err error) {
	return c.resolve(receiver, "")
}

func (c *container) ResolveWithAlias(receiver interface{}, alias string) (err error) {
	return c.resolve(receiver, alias)
}
