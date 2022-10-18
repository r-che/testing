package clone

import (
	"fmt"
	"strings"
	rt "reflect"
)

// StructVerify returns an error if the structure clonning process is not correct.
// It takes the creator function that should create a new instance of the structure, and
// the cloner function that should create a clone of the given argument and return it.
// NOTE: Only exported fields can be verified.
func StructVerify(creator func()any, cloner func(x any) any) error {
	// Make an original value
	orig, err := autoFill(creator())
	if err != nil {
		return fmt.Errorf("cannot autofill original structure: %w", err)
	}

	// And the reference to compare after clone modifications
	ref, err := autoFill(creator())
	if err != nil {
		return fmt.Errorf("cannot autofill reference structure: %w", err)
	}

	// They must be the same
	if !rt.DeepEqual(orig, ref) {
		return fmt.Errorf("newly created and filled structures (original and reference)" +
			" ARE NOT SAME: orig - %#v, ref - %#v", orig, ref)
	}

	// Create clone for each existing field and update the field, check correctness
	for _, field := range structFields(creator()) {
		// Make a clone
		clone := cloner(orig)

		// Update field in the clone
		if err := autoChange(clone, field); err != nil {
			return fmt.Errorf("cannot update field %q in the CLONE: %w", field,  err)
		}
	
		// Compare the original and the reference - they should be the same
		if !rt.DeepEqual(orig, ref) {
			return fmt.Errorf("the ORIGINAL value (%#v) is DIFFERENT from the REFERENCE (%#v)" +
				" after the CLONE FIELD ----> %q <---- has been CHANGED, clone: %#v", orig, ref, field, clone)
		}

		// Compare the clone and the original structure - they should NOT be the same
		if rt.DeepEqual(orig, clone) {
			return fmt.Errorf("CLONE field %q has been UPDATED but the clone is EQUAL the ORIGINAL value: %#v", field, clone)
		}
	}

	// OK
	return nil
}

// autoFill automatically fills the fields of the structure of supported types. It returns
// interface to the filled structure or an error if structure contains fields of unsupported types
func autoFill(si any) (any, error) {
	s := rt.ValueOf(si).Elem()

	// Closure to automate creation of int64 values
	var i64v int64
	i64val := func() int64 {
		i64v++
		return i64v
	}

	// Closure to automate creation of slices of int64
	i64vals:= func() []int64 {
		i64v++
		l := i64v*2
		s := make([]int64, 0, l)
		for i := int64(0); i < l; i++ {
			s = append(s, i64v + i)
		}
		return s
	}

	// Closure to automate creation of slices of strings
	nStrs := 2
	mkStrs := func() []string {
		s := make([]string, 0, nStrs + 1)
		baseChar := fmt.Sprintf("%c", ('a' - 2) + nStrs % ('z' - 'a'))
		for i := 0; i < nStrs; i++ {
			s = append(s, strings.Repeat(baseChar+"_", nStrs))
		}
		nStrs++

		return s
	}

	// Closure to automate creation of map[string]any
	mkMapStrAny := func() map[string]any {
		m := make(map[string]any, nStrs)
		baseChar := fmt.Sprintf("%c", ('a' - 2) + nStrs % ('z' - 'a'))
		for i := 0; i < nStrs; i++ {
			m[strings.Repeat(baseChar+"_", nStrs+i)] = (i+1) * 3 / 2
		}
		nStrs++

		return m
	}

	for i := 0; i < s.NumField(); i++ {
		// Get the i-field
		f := s.Field(i)

		// Check for support types, set some values
		switch f.Type().Kind() {
			case rt.Int64:
				f.Set(rt.ValueOf(i64val()))

			case rt.Map:
				switch f.Interface().(type) {
				case map[string]any:
					f.Set(rt.ValueOf(mkMapStrAny()))
				default:
					return nil, fmt.Errorf("field %q has unsupported type %q", s.Type().Field(i).Name, f.Type())
				}

			case rt.Slice:
				switch f.Interface().(type) {
				case []string:
					f.Set(rt.ValueOf(mkStrs()))
				case []int64:
					f.Set(rt.ValueOf(i64vals()))
				default:
					return nil, fmt.Errorf("field %q has unsupported type %q", s.Type().Field(i).Name, f.Type())
				}

			default:
				return nil, fmt.Errorf("field %q has unsupported type %q", s.Type().Field(i).Name, f.Type().Kind())
		}
	}

	return si, nil
}

// structFields returns a list of field names of the structure specified by si
func structFields(si any) []string {
	var fields []string

	s := rt.ValueOf(si).Elem()
	for i := 0; i < s.NumField(); i++ {
		fields = append(fields, s.Type().Field(i).Name)
	}

	return fields
}

// autoFill automatically changed the fields of the structure of supported types.
// It returns an error if structure contains fields of unsupported types
func autoChange(si any, field string) error {
	s := rt.ValueOf(si).Elem()

	for i := 0; i < s.NumField(); i++ {
		if s.Type().Field(i).Name != field {
			continue
		}

		// Get the current struct's field
		f := s.Field(i)

		switch f.Type().Kind() {
			case rt.Int64:
				// Mult the value to 2
				f.Set(rt.ValueOf(f.Interface().(int64) * 2))
			case rt.Map:
				if m, ok := f.Interface().(map[string]any); ok {
					// Update only one value if exists
					for k, v := range m {
						// Mult the value to 2
						m[k] = v.(int) * 2
						break
					}
				} else {
					return fmt.Errorf("field %q has unsupported type %q", f.Type().Field(i).Name, f.Type())
				}
			case rt.Slice:
				if strSlice, ok := f.Interface().([]string); ok {
					// Concatenate the last value in slice with itself
					strSlice[len(strSlice)-1] += strSlice[len(strSlice)-1]
				} else if i64Slice, ok := f.Interface().([]int64); ok {
					// Mult the last value in slice to 2
					i64Slice[len(i64Slice)-1] *= 2
				} else {
					return fmt.Errorf("field %q has unsupported type %q", f.Type().Field(i).Name, f.Type())
				}
			default:
				return fmt.Errorf("field %q has unsupported type %q", s.Type().Field(i).Name, f.Type().Kind())
		}

		// OK, field found and updated
		return nil
	}

	return fmt.Errorf("field %q was not found in structure", field)
}
