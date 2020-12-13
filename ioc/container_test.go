package ioc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	intProp int
}

type dTestInterface interface {
	GetIntProp() int
}

type dTestStruct struct {
	testStruct *testStruct
}

func (d *dTestStruct) GetIntProp() int {
	if d.testStruct != nil {
		return d.testStruct.intProp
	}

	return -1
}

type dTestTagStruct struct {
	testStruct *testStruct `ioc:"test"`
}

func (d *dTestTagStruct) GetIntProp() int {
	if d.testStruct != nil {
		return d.testStruct.intProp
	}

	return -1
}

func checkMustPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Fatalf("should be panic")
	}
}

func testContainerMustResolve(t *testing.T, cnt Container, v interface{}, opts ...ResolveOption) {
	if err := cnt.Resolve(v, opts...); err != nil {
		t.Fatalf("failed to resolve, err: %v", err)
	}
}

func TestContainer_MustBindSingleton(t *testing.T) {
	t.Run("bind singleton non function resolver", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(&testStruct{}, nil)
	})

	t.Run("bind singleton resolver empty return", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(func() {}, nil)
	})

	t.Run("bind singleton resolver return non pointer or non interface", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(func(bound *testStruct) dTestStruct {
			return dTestStruct{testStruct: bound}
		}, nil)
	})

	t.Run("bind singleton meta not pointer", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(dTestTagStruct{}))
	})

	t.Run("bind singleton meta not implements interface", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&testStruct{}))
	})

	t.Run("bind singleton non function resolver with alias", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(&testStruct{}, WithBindAlias("panic"))
	})

	t.Run("bind singleton pointer struct with/without dependencies", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct })
		cnt.MustBindSingleton(func(bound *testStruct) *dTestStruct {
			return &dTestStruct{testStruct: bound}
		})

		var firstDTest *dTestStruct
		testContainerMustResolve(t, cnt, &firstDTest)

		assert.NotNil(t, firstDTest.testStruct)
		assert.Equal(t, boundStruct, firstDTest.testStruct)
		firstDTest.testStruct = &testStruct{intProp: 2}

		var secondDTest *dTestStruct
		testContainerMustResolve(t, cnt, &secondDTest)

		assert.NotEqual(t, boundStruct, secondDTest.testStruct)
		assert.Equal(t, firstDTest, secondDTest)
	})

	t.Run("bind singleton interface with alias tag dependencies", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct }, WithBindAlias("test"))
		cnt.MustBindSingleton(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&dTestTagStruct{}))

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest)
		assert.Equal(t, boundStruct, firstDTest.(*dTestTagStruct).testStruct)
	})

	t.Run("bind singleton with dependencies outside internal property", func(t *testing.T) {
		cnt := CreateContainer()

		type testOutsideStruct struct {
			internalProps string
		}

		var boundStruct = &testStruct{intProp: 1}
		var outsideStruct = &testOutsideStruct{internalProps: "test"}
		cnt.MustBindSingleton(func() *testOutsideStruct { return outsideStruct })
		cnt.MustBindSingleton(func() *testStruct { return boundStruct }, WithBindAlias("test"))
		cnt.MustBindSingleton(func(bound *testStruct, outside *testOutsideStruct) dTestInterface {
			assert.Equal(t, outsideStruct, outside)

			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&dTestTagStruct{}))

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest)
		assert.Equal(t, boundStruct, firstDTest.(*dTestTagStruct).testStruct)
	})

	t.Run("bind singleton interface with alias and alias tag dependencies", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct }, WithBindAlias("test"))
		cnt.MustBindSingleton(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&dTestTagStruct{}), WithBindAlias("first"))

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest, WithResolveAlias("first"))
		assert.Equal(t, boundStruct, firstDTest.(*dTestTagStruct).testStruct)

		var secondTest dTestInterface
		if err := cnt.Resolve(&secondTest, WithResolveAlias("no_alias")); err != nil {
			assert.Equal(t, ErrAliasNotKnown, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}

func TestContainer_MustBindTransient(t *testing.T) {
	t.Run("bind transient non function resolver", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindTransient(&testStruct{}, nil)
	})

	t.Run("bind transient non function resolver with alias", func(t *testing.T) {
		defer checkMustPanic(t)

		cnt := CreateContainer()
		cnt.MustBindSingleton(&testStruct{}, WithBindAlias("panic"))
	})

	t.Run("bind transient pointer struct", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct })
		cnt.MustBindTransient(func(bound *testStruct) *dTestStruct {
			return &dTestStruct{testStruct: bound}
		})

		var firstDTest *dTestStruct
		testContainerMustResolve(t, cnt, &firstDTest)

		assert.NotNil(t, firstDTest.testStruct)
		assert.Equal(t, boundStruct, firstDTest.testStruct)
		firstDTest.testStruct = &testStruct{intProp: 2}

		var secondDTest *dTestStruct
		testContainerMustResolve(t, cnt, &secondDTest)

		assert.Equal(t, boundStruct, secondDTest.testStruct)
		assert.NotEqual(t, firstDTest, secondDTest)
	})

	t.Run("bind transient interface", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct }, WithBindAlias("test"))
		cnt.MustBindTransient(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&dTestTagStruct{}))

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest)
		firstV := firstDTest.(*dTestTagStruct)
		assert.Equal(t, boundStruct, firstV.testStruct)
		firstV.testStruct = &testStruct{intProp: 2}

		var secondDTest dTestInterface
		testContainerMustResolve(t, cnt, &secondDTest)
		secondV := secondDTest.(*dTestTagStruct)
		assert.Equal(t, boundStruct, secondV.testStruct)
		assert.NotEqual(t, secondV, firstV)
	})

	t.Run("bind transient interface with alias", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindSingleton(func() *testStruct { return boundStruct }, WithBindAlias("test"))
		cnt.MustBindTransient(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, WithBindMeta(&dTestTagStruct{}), WithBindAlias("first"))

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest, WithResolveAlias("first"))
		firstV := firstDTest.(*dTestTagStruct)
		assert.Equal(t, boundStruct, firstV.testStruct)
		firstV.testStruct = &testStruct{intProp: 2}

		var secondTest dTestInterface
		if err := cnt.Resolve(&secondTest, WithResolveAlias("no_alias")); err != nil {
			assert.Equal(t, ErrAliasNotKnown, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}
