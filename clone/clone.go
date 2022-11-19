package clone

import (
	"fmt"
	"reflect"
)

// CreatorFunc defines a function type to create a structure of the tested type
// and return a pointer to it. See [NewStructVerifier] for more details.
type CreatorFunc func() any

// ClonerFunc defines the type of function that takes a pointer to a structure
// of the tested type (created by CreatorFunc) and returns its clone.
// See [NewStructVerifier] for more details.
type ClonerFunc func(x any) any

/*
Changer defines the type of function used in [StructVerifier.Verify] to
automatically
change the field of the appropriate type.

The Changer function must check that the actual data type of the given field is
the type that it can change.

If so, it should perform some correct operations to change the data of the
given field and return true, that will stop the lookup of a proper Changer to
change the field data, after that the verification process will go on to change
the next field.

Otherwise, it must return false to allow the field to be passed to the next
Changer function.

Set of Changer functions supported by default provided by [EmbChangers].

You can define and provide your own Changer functions.

# Provide your own Changer

An example of the Changer function to change a field with type []int:

  // intSliceChanger multiplies each slice element by two
  func intSliceChanger(v reflect.Value) bool {
      // Check that the appropriate type of field is passed
      is, ok := v.Interface().([]int)
      if !ok {
          return false
      }
      // Modify the field value
      for i := range is {
          is[i] *= 2
      }
      return true
  }

Then, you need to pass this function to the verifier using [StructVerifier.AddChangers].
*/
type Changer func(v reflect.Value) bool

/*
Setter defines the type of function used in [StructVerifier.Verify] method to
automatically create a value appropriate for the field of a given type. The
value returned by this function will be assigned to the field of the structure
of the corresponding type.

The Setter function must check that the actual data type of the given field is
the type of value it can create. In this case, it allocates memory for the new
value, fills it with some data valid for that type, then returns the created
value.

Otherwise, it must return nil to allow the field to be passed to the next
Setter function.

Set of Setter functions supported by default provided by [EmbSetters].

Note: for adding your own Setter functions, you must pass a value of type
[SetterCreator], not Setter, to [StructVerifier.AddSetters]. See
[SetterCreator] for details.
*/
type Setter func(v reflect.Value) any

/*
SetterCreator defines the type of function used by the Verify method to create [Setter] functions.

For automatic initialization of the structure fields, two conditions must be satisfied:

  1. The same fields of two different structures of the same type must have the
     same values after initialization - only in this case the equivalence check of
     the original and the reference will make sense.
  2. Meanwhile, different fields of the same type must be filled with different
     data sets. This ensures that cloning errors of the original value will be
     detected. For example, when field2 is added to the structure, the code that
     fills it takes the old field field1, which has the same type, as a source
     instead of the original field2 (Ctrl+C,Ctrl+V programming). If in the original
     field values of both fields (field1 and field2) initialized by the same values,
     then in the clone field2 value will match the value of field2 in the source
     (because field1 and field2 have the same value). Therefore, the error of
     incomplete cloning will not be detected.

To satisfy both conditions, [StructVerifier.AddSetters] takes a function that
creates a Setter function that uses some initializing value defined by the
function that spawned it. For example, the SetterCreator function for fields of
type []int can be like this:

  func intSliceSetter() Setter {
      var initVal int
      return func(v reflect.Value) any {
          if _, ok := v.Interface().([]int); !ok {
              return nil
          }

          initVal++

          l := initVal * 2    // slice length
          s := make([]int, 0, l)
          for i := 0; i < l; i++ {
              s = append(s, initVal + i)
          }

          return s
      }
  }

When filling the original structure, the function intSliceSetter is called,
which returns a Setter function consistently applied to all fields of type
[]int. As the result, the first field of type []int will get value []int{1, 2},
the next field of the same type will get value []int{2, 3, 4, 5}, and so on.

When filling the reference structure, the intSliceSetter function will be
called again to get a Setter function in which the value of the initVal
initializing variable will again be zero, hence - the fields of the reference
structure will have the same values as the original one.
*/
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

/*
NewStructVerifier returns the pointer to the created StructVerifier. It takes
the creator function that creates a new instance of the structure, and the
cloner function, that creates a clone of the given argument and return it.
These functions will be used to create the original, reference and cloned
objects during the verification process.

See [StructVerifier.Verify] for how they are used during verification.
*/
func NewStructVerifier(creator CreatorFunc, cloner ClonerFunc) *StructVerifier {
	return &StructVerifier{
		creator: creator,
		cloner:	cloner,
	}
}

/*
AddChangers adds a user-defined [SetterCreator] function that allows you to
initialize the values of fields with a type not supported by the set of
[Setter] functions provided by [EmbSetters], or to replace them. User-defined
functions added using AddSetters take precedence over embedded Setter functions.

See [Setter] and [SetterCreator] to understand how to create your own Setter
function.
*/
func (sv *StructVerifier) AddSetters(setters ...SetterCreator) *StructVerifier {
	sv.setters = append(sv.setters, setters...)
	return sv
}

/*
AddChangers adds a user-defined [Changer] function that allows you to change
the values of fields with a type not supported by the set of Setter functions
provided by [EmbChangers], or to replace them. User-defined functions added
using AddChangers take precedence over embedded Changer functions.

See [Changer] to understand how to create your own Changer function, and the
Examples section.
*/
func (sv *StructVerifier) AddChangers(changers ...Changer) *StructVerifier {
	sv.changers = append(sv.changers, changers...)
	return sv
}

/*
Verify performs the verification process. It returns an error if the structure
clonning process is not correct.

During verification, the following objects are created and used:

  - Original object - will be created by calling the creator function as a
    clone sample, it will be passed as a parameter to the cloner function
  - Reference object - is created by calling the creator function and is used
    for: 1) checking that the creator function works properly - the original object
    and the reference object are compare before cloning verification, if they are
    equal then creator works correctly; 2) checking that the original object was
    not changed during modifications of the clone object created from it
  - Cloned object - created by calling the cloner function that takes the
    original object as a parameter.

The verification process consists of:

  1. Creation of original and reference objects, compare them with each other -
     they must be equal.
  2. Creation of a clone object from the original object using the cloner function.
  3. Comparison of the original object with the clone - they must be equal.
  4. Automatically change the data of the exported fields of the clone object
     using the Setter functions that match the field types.
  5. Verification that the original object is the same as the reference one -
     it means that modifications of the clone object did not affect the original
     object (or if you have a very bad design, also affected the reference object by
     changing it simultaneously with the original).
  6. Verification that the clone object is different from the original object,
     which should reveal the situation of simultaneous modification of all three
     objects, or incorrect work of Changer-functions.

Verification is considered successful when all the checks are passed.

# Only exported fields cloning can be verified

The reason for this is that all fields of the structure need to be modified for
a full verification. Go, however, forbids changing the values of non-exportable
fields by a code not related to the package of the verified structure.

Your structure can contain non-exported fields, they will be skipped during
verification.

*/
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
