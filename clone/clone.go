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
	Setter func(v rt.Value) any
	// Change supported types of fields
	Changer func(v rt.Value) bool

	// Setters-creator type
	sCreator func() Setter
)

type StructVerifier struct {
	creator	CreatorFunc
	cloner	ClonerFunc

	setters		[]sCreator	// user defined setters
	changers	[]Changer	// user defined changers
}

const initialSeed = 2

//
// Errors
//
type StructVerifierError struct {
	err error
}
func (esv StructVerifierError) Error() string {
	return esv.err.Error()
}
func NewErrSV(format string, args ...any) StructVerifierError {
	return StructVerifierError{fmt.Errorf(format, args...)}
}
type (
	ErrSVOrigFill struct { StructVerifierError }
	ErrSVRefFill struct { StructVerifierError }
	ErrSVRefOrigEqual struct { StructVerifierError }
	ErrSVChange struct { StructVerifierError }
	ErrSVOrigChanged struct { StructVerifierError }
	ErrSVCloneOrigEqual struct { StructVerifierError }
	ErrSVFieldNotFound struct { StructVerifierError }
)

//
// Verifier creator function
//

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

func (sv *StructVerifier) AddChangers(changers ...Changer) *StructVerifier {
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
		return &ErrSVOrigFill{NewErrSV("cannot autofill original structure: %w", err)}
	}

	// And the reference to compare after clone modifications
	ref, err := sv.autoFill()
	if err != nil {
		return &ErrSVRefFill{NewErrSV("cannot autofill reference structure: %w", err)}
	}

	// They must be the same
	if !rt.DeepEqual(orig, ref) {
		return &ErrSVRefOrigEqual{NewErrSV("newly created and filled structures (original and reference)" +
			" ARE NOT SAME: orig - %#v, ref - %#v", orig, ref)}
	}

	// Create clone for each existing field and update the field, check correctness
	for _, field := range structFields(sv.creator()) {
		// Make a clone
		clone := sv.cloner(orig)

		// Update field in the clone
		if err := sv.autoChange(clone, field); err != nil {
			return &ErrSVChange{NewErrSV("cannot update field %q in the CLONE: %w", field,  err)}
		}
	
		// Compare the original and the reference - they should be the same
		if !rt.DeepEqual(orig, ref) {
			return &ErrSVOrigChanged{NewErrSV("the ORIGINAL value (%#v) is DIFFERENT from the REFERENCE (%#v)" +
				" after the CLONE FIELD ----> %q <---- has been CHANGED, clone: %#v", orig, ref, field, clone)}
		}

		// Compare the clone and the original structure - they should NOT be the same
		if rt.DeepEqual(orig, clone) {
			return &ErrSVCloneOrigEqual{NewErrSV(
				"CLONE field %q has been UPDATED but the clone is EQUAL the ORIGINAL value: %#v", field, clone)}
		}
	}

	// OK
	return nil
}

// autoFill automatically creates struct and fills the fields of supported types. It returns
// interface to the filled structure or an error if structure contains fields of unsupported types
func (sv *StructVerifier) autoFill() (any, error) {
	// Create an empty structure instance
	inst := sv.creator()

	// Convert inerface to reflect.Value
	s := rt.ValueOf(inst).Elem()

	// Create new user defined setters to refresh initial values
	uSetters := make([]Setter, 0, len(sv.setters))
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

	return inst, nil
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
	structVal := rt.ValueOf(si).Elem()

	for i := 0; i < structVal.NumField(); i++ {
		if structVal.Type().Field(i).Name != field {
			continue
		}

		// Get the current struct'structVal field
		f := structVal.Field(i)

		// Try to change values using user defined and embedded changers
		for _, changer := range append(sv.changers, embChangers()...) {
			if changer(f) {
				// Ok, field found and updated
				return nil
			}
		}

		// No suitable setter - unsupported type of field
		return &ErrSVChange{NewErrSV("field %q has unsupported type to change - %q",
							structVal.Type().Field(i).Name, f.Type())}
	}

	return &ErrSVFieldNotFound{NewErrSV("field %q was not found in the structure %#v", field, structVal.Interface())}
}

func embSetters() []Setter {
	var i64v int64
	nStrs := int(initialSeed)

	return []Setter {
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

			l := i64v * initialSeed	// slice length
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
			baseChar := fmt.Sprintf("%c", ('a' - initialSeed) + nStrs % ('z' - 'a'))
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

// Embedded changers
func embChangers() []Changer {
		return []Changer{
		// int64 - mult the value to initialSeed (2)
		func(v rt.Value) bool {
			iv, ok := v.Interface().(int64)
			if !ok {
				return false
			}
			v.Set(rt.ValueOf(iv * initialSeed))
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
		// []int64 - mult the last value in the slice to initialSeed (2)
		func(v rt.Value) bool {
			is, ok := v.Interface().([]int64)
			if !ok {
				return false
			}

			is[len(is)-1] *= initialSeed

			return true
		},
		// map[string]any - mult each value to initialSeed (2)
		func(v rt.Value) bool {
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
