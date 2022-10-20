package clone

type _TestConfig struct {
	Int64param	int64
	IntList		[]int
	Int64List	[]int64
	StringList	[]string
	MapVals		map[string]any
	unexported1		bool
	_Unexported2	any
}
func New_TestConfig() *_TestConfig {
	return &_TestConfig{}
}
func (c *_TestConfig) Clone() *_TestConfig {
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
