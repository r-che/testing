package clone

import (
	"fmt"
	"reflect"
)

// CreatorFunc defines a function type to create a structure of the tested type
// and return a pointer to it.
type CreatorFunc func() any
// ClonerFunc defines the type of function that takes a pointer to a structure
// of the tested type (created by CreatorFunc) and returns its clone.
type ClonerFunc func(x any) any
// Setter defines the type of function used in Verify method to automatically
// fill a field of an appropriate type.
type Setter func(v reflect.Value) any
// Changer defines the type of function used in Verify method to automatically
// change the field of the appropriate type.
type Changer func(v reflect.Value) bool
// SetterCreator defines the type of function used by the Verify method to
// create Setter functions.
type SetterCreator func() Setter

type StructVerifier struct {
	creator	CreatorFunc
	cloner	ClonerFunc

	setters		[]SetterCreator	// user defined setters
	changers	[]Changer		// user defined changers
}

//
// Errors
//
type structVerifierError struct {
	err error
}
func (esv structVerifierError) Error() string {
	return esv.err.Error()
}
func newErrSV(format string, args ...any) structVerifierError {
	return structVerifierError{fmt.Errorf(format, args...)}
}
type (
	// ErrSVChange represents an error that occurs when the value of a field in the
	// tested structure cannot be changed.
	ErrSVChange struct { structVerifierError }

	// ErrSVCloneOrigEqual represents an error occurred when the initial value of a cloned
	// structure field was not changed after the Setter function was applied to it.
	ErrSVCloneOrigEqual struct { structVerifierError }

	// ErrSVCloneOrigNotEqual represents an error if the original and the cloned
	// structures are different immediately after creation (before the clone changes).
	ErrSVCloneOrigNotEqual struct { structVerifierError }

	// ErrSVFieldNotFound represents the error which occurs if a clone does not
	// contain the original structure field.
	ErrSVFieldNotFound struct { structVerifierError }

	// ErrSVOrigChanged represents the error occurred when the initial structure
	// (cloning source) was changed after modification of the cloned structure.
	ErrSVOrigChanged struct { structVerifierError }

	// ErrSVOrigFill represents an error, that occurs if the source structure
	// cannot be filled automatically.
	ErrSVOrigFill struct { structVerifierError }

	// ErrSVRefFill represents an error that occurs if the reference structure
	// cannot be automatically filled.
	ErrSVRefFill struct { structVerifierError }

	// ErrSVRefOrigEqual represents an error if the original and the reference
	// structures are different immediately after creation (before the clone changes).
	ErrSVRefOrigEqual struct { structVerifierError }
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

func (sv *StructVerifier) AddSetters(setters ...SetterCreator) *StructVerifier {
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
//
// NOTE: Only exported fields can be verified!
func (sv *StructVerifier) Verify() error {
	// Make an original value
	orig, err := sv.autoFill()
	if err != nil {
		return &ErrSVOrigFill{newErrSV("cannot autofill original structure: %w", err)}
	}

	// And the reference to compare after clone modifications
	ref, err := sv.autoFill()
	if err != nil {
		return &ErrSVRefFill{newErrSV("cannot autofill reference structure: %w", err)}
	}

	// They must be the same
	if !reflect.DeepEqual(orig, ref) {
		return &ErrSVRefOrigEqual{newErrSV("newly created and filled structures (original and reference)" +
			" ARE NOT SAME: orig - %#v, ref - %#v", orig, ref)}
	}

	// Create clone for each existing field and update the field, check correctness
	for _, field := range structFields(sv.creator()) {
		// Make a clone
		clone := sv.cloner(orig)

		// Check that the clone is created correctly - immediately after creation
		// it should be the same as the original
		if !reflect.DeepEqual(orig, clone) {
			return &ErrSVCloneOrigNotEqual{newErrSV("newly created clone is not the same as the original:" +
				" orig - %#v, clone - %#v", orig, clone)}
		}

		// Update field in the clone
		if err := sv.autoChange(clone, field); err != nil {
			return &ErrSVChange{newErrSV("cannot update field %q in the CLONE: %w", field,  err)}
		}
	
		// Compare the original and the reference - they should be the same
		if !reflect.DeepEqual(orig, ref) {
			return &ErrSVOrigChanged{newErrSV("the ORIGINAL value (%#v) is DIFFERENT from the REFERENCE (%#v)" +
				" after the CLONE FIELD ----> %q <---- has been CHANGED, clone: %#v", orig, ref, field, clone)}
		}

		// Compare the clone and the original structure - they should NOT be the same
		if reflect.DeepEqual(orig, clone) {
			return &ErrSVCloneOrigEqual{newErrSV(
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
	s := reflect.ValueOf(inst).Elem()

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
		for _, setter := range append(uSetters, EmbSetters()...) {
			if v := setter(f); v != nil {
				// Set field value to v
				f.Set(reflect.ValueOf(v))
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

	s := reflect.ValueOf(si).Elem()
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
	structVal := reflect.ValueOf(si).Elem()

	for i := 0; i < structVal.NumField(); i++ {
		if structVal.Type().Field(i).Name != field {
			continue
		}

		// Get the current struct'structVal field
		f := structVal.Field(i)

		// Try to change values using user defined and embedded changers
		for _, changer := range append(sv.changers, EmbChangers()...) {
			if changer(f) {
				// Ok, field found and updated
				return nil
			}
		}

		// No suitable setter - unsupported type of field
		return &ErrSVChange{newErrSV("field %q has unsupported type to change - %q",
							structVal.Type().Field(i).Name, f.Type())}
	}

	return &ErrSVFieldNotFound{newErrSV("field %q was not found in the structure %#v", field, structVal.Interface())}
}
