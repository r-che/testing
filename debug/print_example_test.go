package debug

func Example_printSliceDefault() {
	slice := []string{"one", "two", "three"}

	PrintSlice(slice)

	// Output:
	// [#0:one #1:two #2:three]
}

func Example_printSliceTypeLenCap() {
	slice := []string{"one", "two", "three"}

	PrintSlice(slice, PrintType | PrintLenCap)

	// Output:
	// []string(3:3){#0:one #1:two #2:three}
}

func Example_printSliceNil() {
	var nilSlice []any

	PrintSlice(nilSlice)

	// Output:
	// []
}

//nolint:lll
func Example_printSliceGoSyntaxValType() {
	slice := []map[string]int{ {"one": 1}, {"two": 2}, {"three": 3} }

	PrintSlice(slice, PrintGoSyntax, PrintValType)

	// Output:
	// [#0(map[string]int):map[string]int{"one":1} #1(map[string]int):map[string]int{"two":2} #2(map[string]int):map[string]int{"three":3}]
}

func Example_printSliceCommaSepNoSharp() {
	slice := []int{1, 1, 2, 3, 5, 8}

	PrintSlice(slice, PrintCommaSep | PrintNoSharp)
	// Output:
	// [0:1, 1:1, 2:2, 3:3, 4:5, 5:8]
}

func Example_printSliceAllFlags() {
	slice := []int{1, 1, 2, 3, 5, 8, 13}

	PrintSlice(slice, PrintType, PrintCommaSep, PrintNoSharp, PrintGoSyntax, PrintLenCap, PrintValType, PrintValPerLine)

	// Output:
	// []int(7:7){
	//   0(int):1,
	//   1(int):1,
	//   2(int):2,
	//   3(int):3,
	//   4(int):5,
	//   5(int):8,
	//   6(int):13
	// }
}

func Example_printSliceStructs() {
	type point struct { x, y int }
	type eventInfo struct {
		cond	bool
		amount	int
		avg		float32
		descr	string
		pos		point
	}
	slice := []eventInfo{
		{
			cond:	true,
			amount:	5,
			avg:	3.434,
			descr:	"positive condition",
			pos:	point{x: 15, y: 83},
		},
	}

	PrintSlice(slice, PrintGoSyntax)

	// Output:
	// [#0:debug.eventInfo{cond:true, amount:5, avg:3.434, descr:"positive condition", pos:debug.point{x:15, y:83}}]
}
