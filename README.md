# go-simple-ioc 
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/josephsalimin/go-simple-ioc)
![unit-tests](https://github.com/josephsalimin/go-simple-ioc/workflows/unit-tests/badge.svg?branch=master)
![codecov](https://codecov.io/gh/josephsalimin/go-simple-ioc/branch/master/graph/badge.svg)

Simple IoC Container in Golang.

## What is IoC?
Inversion of Control (IoC) is a programming principle which is used to invert controls to achieve loose coupling.
Controls refer to any additional responsibilities a class has, for example the flow of an application or control over the flow of an object creation or dependent object creation and binding (see tutorialsteacher).

As the name imply, IoC inverts the control. Any additional controls (that are not part of a class's responsibilities) are moved 
to other controller that is specifically created to control the flow (usually you can find it in a framework, such as SpringBoot).
Thus, helping in designing loosely coupled classes.

### Dependency Inversion and Dependency Injection

According to Robert Martin, Dependency Inversion has some definitions:
1. High-level modules should not depend on low-level modules. Both should depend on the abstraction.
2. Abstractions should not depend on details. Details should depend on abstractions.

By using abstraction (interface in Golang), dependant class does not need to know the actual implementation and just need to know
which methods that can be used from it. Thus, it is making it easier for us to implements multiple implementation and especially
easier for us to mock the object when we are doing unit-testing.

At the heart of IoC is dependency injection. It allows the creation of dependent objects outside of a class and provides those objects to a class through different ways.
In golang, we can actually get those objects via method injections.

### IoC Container

What this library does is create IoC container that implements IoC principle in golang. The container
will be the injector and will inject any dependencies defined to actual struct using method injections in simple ways.

## Features

### Bind singleton

Bind singleton is kind of normal Bind, but with function and injected dependencies as parameters.

Example can be seen in: [here](./examples/bind_singleton).

### Bind transient

Bind transient is like bind singleton, except for each resolve call, it will create a new instance.

## Caveat

1. Can't bind object with circular dependencies.
2. It uses reflection so may cause slower when serving request. Best to use when initialization.

## Want to contribute?
Feel free to clone this repository and create PR! 😁😁

## References
- [wikipedia](https://en.wikipedia.org/wiki/Inversion_of_control)
- [tutorialsteacher](https://www.tutorialsteacher.com/ioc/inversion-of-control)
