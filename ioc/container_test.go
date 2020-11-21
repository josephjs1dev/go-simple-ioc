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

func testRecover(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatalf("recovered from: %v", r)
	}
}

func testContainerMustResolve(t *testing.T, cnt Container, v interface{}) {
	if err := cnt.Resolve(v); err != nil {
		t.Fatalf("failed to resolve, err: %v", err)
	}
}

func testContainerMustResolveWithAlias(t *testing.T, cnt Container, v interface{}, alias string) {
	if err := cnt.ResolveWithAlias(v, alias); err != nil {
		t.Fatalf("failed to resolve with alias %v, err: %v", alias, err)
	}
}

func TestContainer_MustBind(t *testing.T) {
	defer testRecover(t)

	t.Run("bind non-pointer struct", func(t *testing.T) {
		cnt := CreateContainer()
		cnt.MustBind(&testStruct{intProp: 1})

		var firstTest testStruct
		testContainerMustResolve(t, cnt, &firstTest)
		assert.Equal(t, 1, firstTest.intProp)
		firstTest.intProp = 2

		var secondTest testStruct
		testContainerMustResolve(t, cnt, &secondTest)
		assert.Equal(t, 1, secondTest.intProp)

		assert.NotEqual(t, firstTest.intProp, secondTest.intProp)
	})

	t.Run("bind pointer struct", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBind(&boundStruct)

		var firstTest *testStruct
		testContainerMustResolve(t, cnt, &firstTest)
		assert.Equal(t, boundStruct.intProp, firstTest.intProp)
		firstTest.intProp = 2
		assert.Equal(t, boundStruct.intProp, firstTest.intProp)

		var secondTest *testStruct
		testContainerMustResolve(t, cnt, &secondTest)
		assert.Equal(t, boundStruct.intProp, secondTest.intProp)
	})

	t.Run("bind non-registered struct", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBind(&boundStruct)

		var firstTest testStruct
		if err := cnt.Resolve(&firstTest); err != nil {
			assert.Equal(t, ErrNotRegistered, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}

func TestContainer_MustBindWithAlias(t *testing.T) {
	defer testRecover(t)

	t.Run("bind pointer struct with alias", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindWithAlias(&boundStruct, "first")

		var firstTest *testStruct
		testContainerMustResolveWithAlias(t, cnt, &firstTest, "first")
		assert.Equal(t, boundStruct.intProp, firstTest.intProp)

		var secondTest *testStruct
		if err := cnt.ResolveWithAlias(&secondTest, "no_alias"); err != nil {
			assert.Equal(t, ErrAliasNotKnown, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}

func TestContainer_MustBindSingleton(t *testing.T) {
	defer testRecover(t)

	t.Run("bind singleton pointer struct", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBind(&boundStruct)
		cnt.MustBindSingleton(func(bound *testStruct) *dTestStruct {
			return &dTestStruct{testStruct: bound}
		}, nil)

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

	t.Run("bind singleton interface", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindWithAlias(&boundStruct, "test")
		cnt.MustBindSingleton(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, &dTestTagStruct{})

		var firstDTest dTestInterface
		testContainerMustResolve(t, cnt, &firstDTest)
		assert.Equal(t, boundStruct, firstDTest.(*dTestTagStruct).testStruct)
	})
}

func TestContainer_MustBindSingletonWithAlias(t *testing.T) {
	defer testRecover(t)

	t.Run("bind singleton interface with alias", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindWithAlias(&boundStruct, "test")
		cnt.MustBindSingletonWithAlias(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, &dTestTagStruct{}, "first")

		var firstDTest dTestInterface
		testContainerMustResolveWithAlias(t, cnt, &firstDTest, "first")
		assert.Equal(t, boundStruct, firstDTest.(*dTestTagStruct).testStruct)

		var secondTest dTestInterface
		if err := cnt.ResolveWithAlias(&secondTest, "no_alias"); err != nil {
			assert.Equal(t, ErrAliasNotKnown, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}

func TestContainer_MustBindTransient(t *testing.T) {
	defer testRecover(t)

	t.Run("bind transient pointer struct", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBind(&boundStruct)
		cnt.MustBindTransient(func(bound *testStruct) *dTestStruct {
			return &dTestStruct{testStruct: bound}
		}, nil)

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
		cnt.MustBindWithAlias(&boundStruct, "test")
		cnt.MustBindTransient(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, &dTestTagStruct{})

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
}

func TestContainer_MustBindTransientWithAlias(t *testing.T) {
	defer testRecover(t)

	t.Run("bind transient interface with alias", func(t *testing.T) {
		cnt := CreateContainer()

		var boundStruct = &testStruct{intProp: 1}
		cnt.MustBindWithAlias(&boundStruct, "test")
		cnt.MustBindTransientWithAlias(func(bound *testStruct) dTestInterface {
			return &dTestTagStruct{testStruct: bound}
		}, &dTestTagStruct{}, "first")

		var firstDTest dTestInterface
		testContainerMustResolveWithAlias(t, cnt, &firstDTest, "first")
		firstV := firstDTest.(*dTestTagStruct)
		assert.Equal(t, boundStruct, firstV.testStruct)
		firstV.testStruct = &testStruct{intProp: 2}

		var secondTest dTestInterface
		if err := cnt.ResolveWithAlias(&secondTest, "no_alias"); err != nil {
			assert.Equal(t, ErrAliasNotKnown, err)
		} else {
			t.Fatalf("should be failed to resolve")
		}
	})
}
