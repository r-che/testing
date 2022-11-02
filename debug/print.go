package debug

import "fmt"

type PrintFlag byte
const (
	PrintType		=	PrintFlag(1) << iota
	PrintCommaSep
	PrintNoSharp
	PrintGoSyntax
	PrintLenCap
)

func PrintSlice[T any](slice []T, flagsVariadic ...PrintFlag) {
	// Open/closed braces
	ob, cb := "[", "]"

	// Get flags if specified
	var flags PrintFlag
	if len(flagsVariadic) != 0 {
		flags = flagsVariadic[0]
	}

	// Is printing of slice type required?
	if flags & PrintType != 0 {
		// Print slice type
		fmt.Printf("%T", slice)
		// Replace open/closed braces to make Go-like output
		ob, cb = "{", "}"
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
	outFmt += "%d:%"
	if flags & PrintGoSyntax != 0 {
		outFmt += "#"
	}
	outFmt += "v"

	// Print open brace
	fmt.Print(ob)
	for i, v := range slice {
		fmt.Printf(outFmt, i, v)
		if i != len(slice) - 1 {
			if flags & PrintCommaSep != 0 {
				fmt.Print(",")
			}
			fmt.Print(" ")
		}
	}
	// Print closed brace
	fmt.Println(cb)
}
