package clone

import (
	"fmt"
	"reflect"
)

type Config struct {
	Int64param	int64
	IntList		[]int
	Int64List	[]int64
	StringList	[]string
	MapVals		map[string]any
}
func NewConfig() *Config {
	return &Config{}
}
func (c *Config) Clone() *Config {
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

func ExampleStructVerify() {
	sv := NewStructVerifier(
		// Creator function
		func() any { return NewConfig() },
		// Cloner function
		func(x any) any {
			c, ok := x.(*Config)
			if ! ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *Config", x))
			}
			return c.Clone()
	}).
		AddSetters(intSliceSetter).
		AddChangers(intSliceChanger)

	err := sv.Verify()

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Verification successful")
	}
	// Output:
	// Verification successful
}

func intSliceSetter() setter {
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

	is [len(is)-1] *= 2

	return true
}
