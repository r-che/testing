package debug

import "fmt"

type PrintFlag byte
const PrintNoFlags	=	0
const (
	PrintType		=	PrintFlag(1) << iota
	PrintCommaSep
	PrintNoSharp
	PrintGoSyntax
	PrintLenCap
	PrintValType
)

func PrintSlice[T any](slice []T, flagsVariadic ...PrintFlag) {
	// Open/closed braces
	obr, cbr := "[", "]"

	// Get flags if specified
	flags := mergeFlags(flagsVariadic)

	// Is printing of slice type required?
	if flags & PrintType != 0 {
		// Print slice type
		fmt.Printf("%T", slice)
		// Replace open/closed braces to make Go-like output
		obr, cbr = "{", "}"
	}

	// Is printing of length and capacity required?
	if flags & PrintLenCap != 0 {
		fmt.Printf("(%d:%d)", len(slice), cap(slice))
	}

	// Output format
	outFmt := ""
	if flags & PrintNoSharp == 0 {
		outFmt += "#"
	}
	outFmt += "%d%s:%"
	if flags & PrintGoSyntax != 0 {
		outFmt += "#"
	}
	outFmt += "v"

	// Print open brace
	fmt.Print(obr)
	for i, v := range slice {
		// Type of value string
		var valType string
		// Is it required?
		if flags & PrintValType != 0 {
			// Set value
			valType = fmt.Sprintf("(%T)", v)
		}

		fmt.Printf(outFmt, i, valType, v)
		if i != len(slice) - 1 {
			if flags & PrintCommaSep != 0 {
				fmt.Print(",")
			}
			fmt.Print(" ")
		}
	}
	// Print closed brace
	fmt.Println(cbr)
}

func mergeFlags(flagsVariadic []PrintFlag) PrintFlag {
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
		var flags PrintFlag
		for _, flag := range flagsVariadic {
			flags |= flag
		}

		return flags
	}
}
