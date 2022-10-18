package clone

import (
	"fmt"
)

type Config struct {
	Int64param	int64
	Int64list	[]int64
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

	rv.Int64list = make([]int64, len(c.Int64list))
	copy(rv.Int64list, c.Int64list)

	rv.StringList = make([]string, len(c.StringList))
	copy(rv.StringList, c.StringList)

	rv.MapVals = make(map[string]any, len(c.MapVals))
	for k, v := range c.MapVals {
		rv.MapVals[k] = v
	}

	return &rv
}

func ExampleStructVerify() {
	err := StructVerify(
		// Creator function
		func() any { return NewConfig() },
		// Cloner function
		func(x any) any {
			c, ok := x.(*Config)
			if ! ok {
				panic(fmt.Sprintf("unsupported type to clone: got - %T, want - *Config", x))
			}
			return c.Clone()
		},
	)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("Verification successful")
	}
	// Output:
	// Verification successful
}
