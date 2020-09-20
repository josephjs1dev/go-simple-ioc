package ioc

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotRegistered = errors.New("information is not registered to container")
	ErrAliasNotKnown = errors.New("alias is not known")
)

// Container ...
type Container interface {
	Bind(instance interface{})
	BindWithAlias(instance interface{}, alias string)
	Resolve(instance interface{}) (err error)
	ResolveWithAlias(instance interface{}, alias string) (err error)
}

type dependency [2]string

type binder struct {
	// singleton flag
	singleton bool
	// metadata of the resolver
	meta interface{}
	// actual instance
	instance interface{}
	// dependencies is a list of dependency from the implementation
	dependencies []dependency
	// resolver function
	resolver interface{}
}

// Implementation of Container interface
type container struct {
	// Map of string to map of string interface
	// First key is the type (can be interface or struct) while second key is alias (default is default key) to the implementation
	cnt map[string]map[string]binder
}

func (c *container) bind(instance interface{}, alias string) {
	instanceType := reflect.TypeOf(instance)

	// Use panic as we don't allow function when binding.
	if instanceType.Kind() == reflect.Func {
		panic("instance must not be a function")
	}

	instanceType = instanceType.Elem()
	label := fmt.Sprintf("%s.%s", instanceType.PkgPath(), instanceType.Name())
	binderLabel := "default"
	if alias != "" {
		binderLabel = alias
	}

	if v, ok := c.cnt[label]; !ok {
		c.cnt[label] = map[string]binder{binderLabel: {singleton: true, instance: instance}}
	} else {
		v[binderLabel] = binder{singleton: true, instance: instance}
	}

	fmt.Println(c.cnt)
}

// Bind does the binding of actual object (struct, not interface) without any alias.
func (c *container) Bind(instance interface{}) {
	c.bind(instance, "")
}

// BindWithAlias does the binding of actual object (struct, not interface) but with an alias.
func (c *container) BindWithAlias(instance interface{}, alias string) {
	c.bind(instance, alias)
}

func (c *container) bindFunc(resolver interface{}, meta interface{}, singleton bool, alias string) {
	resolverType := reflect.TypeOf(resolver)

	// Panic when resolver is not function
	if resolverType.Kind() != reflect.Func {
		panic("resolver must be a function")
	}
}

func (c *container) BindSingleton(resolver interface{}, meta interface{}) {}

func (c *container) BindSingletonWithAlias(resolver interface{}, meta interface{}, alias string) {}

func (c *container) BindTransient(resolver interface{}, meta interface{}) {}

func (c *container) BindTransientWithAlias(resolver interface{}, meta interface{}, alias string) {}

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
	receiverType := reflect.TypeOf(receiver).Elem()
	label := fmt.Sprintf("%s.%s", receiverType.PkgPath(), receiverType.Name())
	binderLabel := "default"
	if alias != "" {
		binderLabel = alias
	}

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
