package debug

import "fmt"

// PrintFlags is a set of flags that configure the Print* functions behavior.
type PrintFlags uint32

// Is method returns true if pf contains any flag of the flagSet.
func (pf PrintFlags) Is(flagsSet PrintFlags) bool {
	return pf & flagsSet != 0
}

// Not method returns true if pf has no flags from the flagSet.
func (pf PrintFlags) Not(flagsSet PrintFlags) bool {
	return pf & flagsSet == 0
}

// [PrintFlags] configure the output format of Print* functions.
const (
	PrintNoFlags	=	PrintFlags(1) << iota
	PrintType		// print type of the argument before the actual argument's content
	PrintCommaSep	// print commas as a content element separator
	PrintNoSharp	// disables print # chbefore the ordinal number of the items
	PrintGoSyntax	// enables Go-syntax style output of argument elements
	PrintLenCap		// print of the length and capacity of the argument before the actual content
	PrintValType	// print the type of each element before print the element's content
	PrintValPerLine	// print one element per line
)

/*
PrintSlice outputs a slice of type T (see [Go generics]). The flagsVariadic parameter determines
the output format and can be a bitmask:
  PrintSlice(slice, debug.PrintType|debug.PrintCommaSep)
or a separately defined argument list:
  PrintSlice(slice, debug.PrintType, debug.PrintCommaSep)

[Go generics]: https://go.dev/blog/intro-generics

By default, PrintSlice output is similar to [fmt.Println] output, but each item
is preceded by its ordinal number, denoted by #, and separated from the item
value by a colon. The output is terminated with a newline character.

For example,

  ints := []int{1, 2, 3, 4}
  debug.PrintSlice(ints)
  
  strs := []string{"one", "two", "three", "four"}
  debug.PrintSlice(strs)

will produce:

  [#0:1 #1:2 #2:3 #3:4]
  [#0:one #1:two #2:three #3:four]

See more examples in the Examples section.

*/
func PrintSlice[T any](slice []T, flagsVariadic ...PrintFlags) {
	// Open/closed braces
	obr, cbr := "[", "]"

	// Get flags if specified
	flags := mergeFlags(flagsVariadic)

	// Is printing of slice type required?
	if flags.Is(PrintType) {
		// Print slice type
		fmt.Printf("%T", slice)
		// Replace open/closed braces to make Go-like output
		obr, cbr = "{", "}"
	}

	// Is printing of length and capacity required?
	if flags.Is(PrintLenCap) {
		fmt.Printf("(%d:%d)", len(slice), cap(slice))
	}

	// Output format
	outFmt := itemFmt(flags)

	// Print open brace
	fmt.Print(obr)

	// Is only one value per line to be printed?
	if flags.Is(PrintValPerLine) {
		// Print new line before the first item
		fmt.Println()
	}

	// Output items
	printSliceItems(outFmt, slice, flags)

	// Print closed brace
	fmt.Println(cbr)
}

func itemFmt(flags PrintFlags) string {
	// Output format
	outFmt := ""

	// Is only one value per line to be printed?
	if flags.Is(PrintValPerLine) {
		// Need to add indentation (2 spaces)
		outFmt += "  "
	}

	// Is printing sharp has not disabled?
	if flags.Not(PrintNoSharp) {
		// Append sharp sign
		outFmt += "#"
	}

	// Appnd position, value type specificator and colon before the value
	outFmt += "%d%s:"

	// Is Go-syntax required in output?
	if flags.Is(PrintGoSyntax) {
		// Append alternative value output format
		outFmt += "%#v"
	} else {
		// Append default value output format
		outFmt += "%v"
	}

	return outFmt
}

func printSliceItems[T any](outFmt string, slice []T, flags PrintFlags) {
	// Items divider
	var iDiv string
	if flags.Is(PrintValPerLine) {
		// Use new line as items separator
		iDiv = "\n"

		// Also need to print new line at end of the output
		defer fmt.Println()
	} else {
		// Use space as items separator
		iDiv = " "
	}

	for i, v := range slice {
		// Type of value string
		var valType string
		// Is it required?
		if flags.Is(PrintValType) {
			// Set value
			valType = fmt.Sprintf("(%T)", v)
		}

		fmt.Printf(outFmt, i, valType, v)

		if i != len(slice) - 1 {
			if flags.Is(PrintCommaSep) {
				fmt.Print(",")
			}
			fmt.Print(iDiv)
		}
	}
}

func mergeFlags(flagsVariadic []PrintFlags) PrintFlags {
	switch len(flagsVariadic) {
	// No flags
	case 0:
		// Return empty flags value
		return PrintNoFlags

	// Flags provided
	case 1:
		// Return flags value from the first element of flagsVariadic
		return flagsVariadic[0]

	// Merge flags from all items of flagsVariadic
	default:
		var flags PrintFlags
		for _, flag := range flagsVariadic {
			flags |= flag
		}

		return flags
	}
}
