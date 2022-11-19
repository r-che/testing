package clone

import (
	"fmt"
	"strings"
	"reflect"
)

const initialSeed = 2

//nolint:cyclop	// In fact, there are no cyclops there
func EmbSetters() []Setter {
	var i64v int64
	var intVal int
	nStrs := int(initialSeed)

	return []Setter {
		// int
		func(v reflect.Value) any {
			if _, ok := v.Interface().(int); !ok {
				return nil
			}

			intVal++

			return intVal
		},

		// int64
		func(v reflect.Value) any {
			if _, ok := v.Interface().(int64); !ok {
				return nil
			}

			i64v++

			return i64v
		},

		// []int
		func(v reflect.Value) any {
			if _, ok := v.Interface().([]int); !ok {
				return nil
			}

			intVal++

			l := intVal * initialSeed	// slice length
			s := make([]int, 0, l)
			for i := 0; i < l; i++ {
				s = append(s, intVal + i)
			}

			return s
		},

		// []int64
		func(v reflect.Value) any {
			if _, ok := v.Interface().([]int64); !ok {
				return nil
			}

			i64v++

			l := i64v * initialSeed	// slice length
			s := make([]int64, 0, l)
			for i := int64(0); i < l; i++ {
				s = append(s, i64v + i)
			}

			return s
		},

		// []string
		func(v reflect.Value) any {
			if _, ok := v.Interface().([]string); !ok {
				return nil
			}

			s := make([]string, 0, nStrs + 1)
			baseChar := fmt.Sprintf("%c", ('a' - initialSeed) + nStrs % ('z' - 'a'))
			for i := 0; i < nStrs; i++ {
				s = append(s, strings.Repeat(baseChar+"_", nStrs))
			}
			nStrs++

			return s
		},

		// map[string]any
		func(v reflect.Value) any {
			if _, ok := v.Interface().(map[string]any); !ok {
				return nil
			}

			m := make(map[string]any, nStrs)
			baseChar := fmt.Sprintf("%c", ('a' - initialSeed) + nStrs % ('z' - 'a'))
			for i := 0; i < nStrs; i++ {
				//nolint:gomnd	// Yes, some kind of pseudo-random generation magic here
				m[strings.Repeat(baseChar+"_", nStrs+i)] = (i+1) * 3 / 2
			}
			nStrs++

			return m
		},
	}
}

func EmbChangers() []Changer {
	return []Changer{
		// int - mult the value to initialSeed (2)
		func(v reflect.Value) bool {
			iv, ok := v.Interface().(int)
			if !ok {
				return false
			}
			v.Set(reflect.ValueOf(iv * initialSeed))
			return true
		},

		// int64 - mult the value to initialSeed (2)
		func(v reflect.Value) bool {
			iv, ok := v.Interface().(int64)
			if !ok {
				return false
			}
			v.Set(reflect.ValueOf(iv * initialSeed))
			return true
		},

		// []int - mult the last value in the slice to initialSeed (2)
		func(v reflect.Value) bool {
			is, ok := v.Interface().([]int)
			if !ok {
				return false
			}

			is[len(is)-1] *= initialSeed

			return true
		},

		// []int64 - mult the last value in the slice to initialSeed (2)
		func(v reflect.Value) bool {
			is, ok := v.Interface().([]int64)
			if !ok {
				return false
			}

			is[len(is)-1] *= initialSeed

			return true
		},

		// []string - concatenate the last value in the slice with itself
		func(v reflect.Value) bool {
			ss, ok := v.Interface().([]string)
			if !ok {
				return false
			}

			ss[len(ss)-1] += ss[len(ss)-1]

			return true
		},

		// map[string]any - mult each value to initialSeed (2)
		func(v reflect.Value) bool {
			m, ok := v.Interface().(map[string]any)
			if !ok {
				return false
			}

			// Update only one random value if exists
			for k, v := range m {
				//nolint:forcetypeassert // Mult the value to initialSeed (2)
				m[k] = v.(int) * initialSeed
				break
			}

			return true
		},
	}
}
