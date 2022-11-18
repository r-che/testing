package clone

import (
	"fmt"
	"reflect"
)

func ExampleStructVerifierSuccess() {
	//
	// Do successful verification
	//
	verifier := NewStructVerifier(
		// Creator function
		func() any { return NewTestConfig() },
		// Cloner function
		func(x any) any {
			c, ok := x.(*_TestConfig)
			if ! ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *_TestConfig", x))
			}
			// Call clone function
			return c.Clone()
	}).
		AddSetters(intSliceSetter).
		AddChangers(intSliceChanger)

	// Run verification
	err := verifier.Verify()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Verification successful\n")
	}

	// Output:
	// Verification successful
}

func ExampleStructVerifierFail() {
	//
	// Do unsuccessful verification
	//
	verifier := NewStructVerifier(
		// Creator function
		func() any { return NewTestConfig() },
		// Cloner function
		func(x any) any {
			c, ok := x.(*_TestConfig)
			if ! ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *_TestConfig", x))
			}
			// XXX Create a copy of the original WITHOUT the actual cloning operation. This will
			// XXX modify the slices in the original structure after changing the fake-clone
			rv := *c
			return &rv
		},
	).
		AddSetters(intSliceSetter).
		AddChangers(intSliceChanger)

	// Run verification
	err := verifier.Verify()
	if err == nil {
		fmt.Println("ERROR: verifier did not catch the original structure modification")
	} else {
		fmt.Printf("Got expected error: %T\n", err)
	}

	// Output:
	// Got expected error: *clone.ErrSVOrigChanged
}

func intSliceSetter() Setter {
	var iv int
	return func(v reflect.Value) any {
		if _, ok := v.Interface().([]int); !ok {
			return nil
		}

		iv++

		l := iv*2	// slice length
		s := make([]int, 0, l)
		for i := 0; i < l; i++ {
			s = append(s, iv + i)
		}

		return s
	}
}

// []int - mult the last value in the slice to 2
func intSliceChanger(v reflect.Value) bool {
	is, ok := v.Interface().([]int)
	if !ok {
		return false
	}

	is[len(is)-1] *= 2

	return true
}
