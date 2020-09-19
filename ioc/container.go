package ioc

import (
	"fmt"
	"reflect"
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

	if instanceType.Kind() == reflect.Ptr {
		instanceType = instanceType.Elem()
	}

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

func (c *container) BindSingletonWithAlias(resolver interface{}, meta interface{}, alias string) {

}

func (c *container) BindTransient() {}

func (c *container) Resolve(instance interface{}) (err error) {

	return nil
}

func (c *container) ResolveWithAlias(instance interface{}, alias string) (err error) {

	return nil
}
