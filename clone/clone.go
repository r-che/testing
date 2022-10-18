package clone

import (
	"fmt"
	"strings"
	rt "reflect"
)

type (
	CreatorFunc func() any
	ClonerFunc func(x any) any

	// Set supported types of fields
	setter func(v rt.Value) any
	// Change supported types of fields
	changer func(v rt.Value) bool

	// Setters-creator type
	sCreator func() setter
)

type StructVerifier struct {
	creator	CreatorFunc
	cloner	ClonerFunc

	setters		[]sCreator	// user defined setters
	changers	[]changer	// user defined changers
}

func NewStructVerifier(creator CreatorFunc, cloner ClonerFunc) *StructVerifier {
	return &StructVerifier{
		creator: creator,
		cloner:	cloner,
	}
}

func (sv *StructVerifier) AddSetters(setters ...sCreator) *StructVerifier {
	sv.setters = append(sv.setters, setters...)
	return sv
}

func (sv *StructVerifier) AddChangers(changers ...changer) *StructVerifier {
	sv.changers = append(sv.changers, changers...)
	return sv
}

// Verify returns an error if the structure clonning process is not correct.
// It takes the creator function that should create a new instance of the structure, and
// the cloner function that should create a clone of the given argument and return it.
// NOTE: Only exported fields can be verified!
func (sv *StructVerifier) Verify() error {
	// Make an original value
	orig, err := sv.autoFill()
	if err != nil {
		return fmt.Errorf("cannot autofill original structure: %w", err)
	}

	// And the reference to compare after clone modifications
	ref, err := sv.autoFill()
	if err != nil {
		return fmt.Errorf("cannot autofill reference structure: %w", err)
	}

	// They must be the same
	if !rt.DeepEqual(orig, ref) {
		return fmt.Errorf("newly created and filled structures (original and reference)" +
			" ARE NOT SAME: orig - %#v, ref - %#v", orig, ref)
	}

	// Create clone for each existing field and update the field, check correctness
	for _, field := range structFields(sv.creator()) {
		// Make a clone
		clone := sv.cloner(orig)

		// Update field in the clone
		if err := sv.autoChange(clone, field); err != nil {
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

// autoFill automatically creates struct and fills the fields of supported types. It returns
// interface to the filled structure or an error if structure contains fields of unsupported types
func (sv *StructVerifier) autoFill() (any, error) {
	// Create empty structure instance
	si := sv.creator()

	// Convert inerface to reflect.Value
	s := rt.ValueOf(si).Elem()

	// Create new user defined setters to refresh initial values
	uSetters := make([]setter, 0, len(sv.setters))
	for _, mkSetter := range sv.setters {
		uSetters = append(uSetters, mkSetter())
	}

	for i := 0; i < s.NumField(); i++ {
		// Get the i-field
		f := s.Field(i)
		name := s.Type().Field(i).Name

		// Filter unexported fields
		if c := name[0]; c == '_' || (c >= 'a' && c <= 'z') {
			// Skip this field
			continue
		}

		// Try to set values using user defined and embedded setters
		for _, setter := range append(uSetters, embSetters()...) {
			if v := setter(f); v != nil {
				// Set field value to v
				f.Set(rt.ValueOf(v))
				// Go to next field
				goto nextField
			}
		}

		// No suitable setter - unsupported type of field
		return nil, fmt.Errorf("field %q has unsupported type to set - %q", name, f.Type())

		nextField:
	}

	return si, nil
}

// structFields returns a list of field names of the structure specified by si
func structFields(si any) []string {
	var fields []string

	s := rt.ValueOf(si).Elem()
	for i := 0; i < s.NumField(); i++ {
		// Filter unexported fields
		name := s.Type().Field(i).Name
		if c := name[0]; c == '_' || (c >= 'a' && c <= 'z') {
			// Skip this field
			continue
		}
		fields = append(fields, name)
	}

	return fields
}

// autoFill automatically changed the fields of the structure of supported types.
// It returns an error if structure contains fields of unsupported types
func (sv *StructVerifier) autoChange(si any, field string) error {
	s := rt.ValueOf(si).Elem()

	for i := 0; i < s.NumField(); i++ {
		if s.Type().Field(i).Name != field {
			continue
		}

		// Get the current struct's field
		f := s.Field(i)

		// Try to change values using user defined and embedded changers
		for _, changer := range append(sv.changers, embChangers...) {
			if changer(f) {
				// Ok, field found and updated
				return nil
			}
		}

		// No suitable setter - unsupported type of field
		return fmt.Errorf("field %q has unsupported type to change - %q", s.Type().Field(i).Name, f.Type())
	}

	return fmt.Errorf("field %q was not found in structure", field)
}

func embSetters() []setter {
	var i64v int64
	var nStrs int = 2

	return []setter {
		// int64
		func(v rt.Value) any {
			if _, ok := v.Interface().(int64); !ok {
				return nil
			}

			i64v++

			return i64v
		},

		// []int64
		func(v rt.Value) any {
			if _, ok := v.Interface().([]int64); !ok {
				return nil
			}

			i64v++

			l := i64v*2	// slice length
			s := make([]int64, 0, l)
			for i := int64(0); i < l; i++ {
				s = append(s, i64v + i)
			}

			return s
		},

		// []string
		func(v rt.Value) any {
			if _, ok := v.Interface().([]string); !ok {
				return nil
			}

			s := make([]string, 0, nStrs + 1)
			baseChar := fmt.Sprintf("%c", ('a' - 2) + nStrs % ('z' - 'a'))
			for i := 0; i < nStrs; i++ {
				s = append(s, strings.Repeat(baseChar+"_", nStrs))
			}
			nStrs++

			return s
		},

		// map[string]any
		func(v rt.Value) any {
			if _, ok := v.Interface().(map[string]any); !ok {
				return nil
			}

			m := make(map[string]any, nStrs)
			baseChar := fmt.Sprintf("%c", ('a' - 2) + nStrs % ('z' - 'a'))
			for i := 0; i < nStrs; i++ {
				m[strings.Repeat(baseChar+"_", nStrs+i)] = (i+1) * 3 / 2
			}
			nStrs++

			return m
		},
	}
}

// Embedded changers
var embChangers = []changer{
	// int64 - mult the value to 2
	func(v rt.Value) bool {
		iv, ok := v.Interface().(int64)
		if !ok {
			return false
		}
		v.Set(rt.ValueOf(iv * 2))
		return true
	},
	// []string - concatenate the last value in the slice with itself
	func(v rt.Value) bool {
		ss, ok := v.Interface().([]string)
		if !ok {
			return false
		}

		ss[len(ss)-1] += ss[len(ss)-1]

		return true
	},
	// []int64 - mult the last value in the slice to 2
	func(v rt.Value) bool {
		is, ok := v.Interface().([]int64)
		if !ok {
			return false
		}

		is [len(is)-1] *= 2

		return true
	},
	// map[string]any - mult each value to 2
	func(v rt.Value) bool {
		m, ok := v.Interface().(map[string]any)
		if !ok {
			return false
		}

		// Update only one random value if exists
		for k, v := range m {
			// Mult the value to 2
			m[k] = v.(int) * 2
			break
		}

		return true
	},
}
