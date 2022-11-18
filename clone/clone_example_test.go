package clone

import (
	"fmt"
	"reflect"
)

func Example_verifySuccess() {
	// Define struct type that need to be tested
	type myStruct struct {
		Int64		int64
		Ints		[]int
		Map			map[string]any
	}

	// Create cloner-function that creates a clone of the myStruct type structure.
	// Note: it can also be a structure method, like: func (ms *myStruct) Clone() *myStruct
	goodCloner := func(orig *myStruct) *myStruct {
		// Make a simple copy of the original structure
		rv := *orig

		// XXX Further, we need to clone all complex fields (slices, maps, etc...)

		// Clone slice with integer using standard copy function
		rv.Ints = make([]int, len(orig.Ints))
		copy(rv.Ints, orig.Ints)

		// Copy all key-value pairs from the map one by one
		rv.Map = make(map[string]any, len(orig.Map))
		for k, v := range orig.Map{
			rv.Map[k] = v
		}

		// Clone process done, return the created clone
		return &rv
	}

	// intSliceSetter creates a Setter function that will used to fill fields with the type []int
	intSliceSetter := func() Setter {
		var nextIntVal int
		return func(v reflect.Value) any {
			if _, ok := v.Interface().([]int); !ok {
				return nil
			}
			nextIntVal++

			newLen := nextIntVal * 2	// slice length
			s := make([]int, 0, newLen)
			for i := 0; i < newLen; i++ {
				s = append(s, nextIntVal + i)
			}

			return s
		}
	}

	// Provide user-defined changer function for []int fields
	// intSliceChanger multiplies the last value in the slice []int by 2
	intSliceChanger := func(v reflect.Value) bool {
		is, ok := v.Interface().([]int)
		if !ok {
			return false
		}
		is[len(is)-1] *= 2
		return true
	}


	// Create a new StructVerifier, you must provide the correct creator and cloning functions
	verifier := NewStructVerifier(
		// Creator function - just return new empty structure
		func() any { return &myStruct{} },
		// Cloner function
		func(x any) any {
			// Need to check that the given argument has the correct type
			v, ok := x.(*myStruct)
			if !ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *myStruct", x))
			}
			// Call clone function to return clone of it
			return goodCloner(v)
	}).
		// Add setter function for the []int type
		AddSetters(intSliceSetter).
		// Add user-defined changer for the []int type
		AddChangers(intSliceChanger)

	// Perform the verification, expecting it to be successful
	err := verifier.Verify()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Verification successful\n")
	}

	// Output:
	// Verification successful
}

func Example_verifyUnsuccessful() {
	// Define struct type that need to be tested
	type myStruct struct {
		Int64		int64
		Map			map[string]any
	}

	// Creation of buggy cloner function that creates a clone of myStruct type structure
	// without actually copying values of Map field
	//
	// Note: it can also be a structure method, like: func (ms *myStruct) Clone() *myStruct
	buggyCloner := func(orig *myStruct) *myStruct {
		// Make a simple copy of the original structure
		rv := *orig

		// XXX We do NOT copy complex fields to make a bug!
		// Instead, return the created clone immediately
		return &rv
	}

	// Create a new StructVerifier, you must provide the correct creator and cloning functions
	verifier := NewStructVerifier(
		// Creator function - just return new empty structure
		func() any { return &myStruct{} },
		// Cloner function
		func(x any) any {
			// Need to check that the given argument has the correct type
			v, ok := x.(*myStruct)
			if !ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *myStruct", x))
			}
			// Call clone function to return clone of it
			return buggyCloner(v)
	})

	// Perform the verification, expecting it to be unsuccessful
	err := verifier.Verify()
	if err == nil {
		fmt.Println("ERROR: verifier did not catch the original structure modification")
	} else {
		fmt.Printf("Got expected error: %T\n", err)
	}

	// Output:
	// Got expected error: *clone.ErrSVOrigChanged
}
