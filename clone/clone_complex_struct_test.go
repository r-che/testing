package clone

import (
	"fmt"
	"reflect"
)

// testComplexStruct is an example of a structure including fields
// of complex types - slices and map and unused unexported fields
type testComplexStruct struct {
	Int64param	int64
	IntList		[]int
	Int64List	[]int64
	StringList	[]string
	MapVals		map[string]any
	// XXX The following fields are not exported and cannot be verified
	unexported1		bool	//nolint:unused	// required for testing
	_Unexported2	any		//nolint:unused	// required for testing
}

func newTestComplexStruct() *testComplexStruct {
	return &testComplexStruct{}
}

// Clone correctly clones the object on which it is called,
// including copying fields of complex types
func (c *testComplexStruct) Clone() *testComplexStruct {
	// Create a simple copy of the configuration
	rv := *c

	//
	// Need to copy all complex fields (slices, maps)
	//

	rv.IntList = make([]int, len(c.IntList))
	copy(rv.IntList, c.IntList)

	rv.Int64List = make([]int64, len(c.Int64List))
	copy(rv.Int64List, c.Int64List)

	rv.StringList = make([]string, len(c.StringList))
	copy(rv.StringList, c.StringList)

	rv.MapVals = make(map[string]any, len(c.MapVals))
	for k, v := range c.MapVals {
		rv.MapVals[k] = v
	}

	return &rv
}

// intSliceSetter creates a Setter function that will used to fill fields with the type []int
func intSliceSetter() Setter {
	var iv int
	return func(v reflect.Value) any {
		if _, ok := v.Interface().([]int); !ok {
			return nil
		}

		iv++

		l := iv*2       // slice length
		s := make([]int, 0, l)
		for i := 0; i < l; i++ {
			s = append(s, iv + i)
		}

		return s
	}
}

// intSliceChanger multiplies the last value in the slice []int by 2
func intSliceChanger(v reflect.Value) bool {
	is, ok := v.Interface().([]int)
	if !ok {
		return false
	}

	is[len(is)-1] *= 2

	return true
}

// The code below demonstrates how to use the verifier for the Clone method on
// the testComplexStruct type, containing fields with complex types like slice and map
func Example_verifySuccessComplex() {
	// Create verifier
	sv := NewStructVerifier(
	    // Creator function
	    func() any { return newTestComplexStruct() },
	    // Cloner function
	    func(x any) any {
	        if c, ok := x.(*testComplexStruct); ok {
	            return c.Clone()
	        }
	        panic(fmt.Sprintf("unsupported type: got - %T, want - *Config", x))
	}).
		// Add our custom setter for the []int type
		AddSetters(intSliceSetter).
		// Add our custom changer for the []int type
		AddChangers(intSliceChanger)

	// Perform the verification, expecting it to be successful
	if err := sv.Verify(); err != nil {
	    fmt.Printf("verification error: %v", err)
	} else {
		fmt.Printf("Verification successful\n")
	}

	// Output:
	// Verification successful
}
