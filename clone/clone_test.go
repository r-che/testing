package clone

import (
	"testing"
	"reflect"
	"errors"
)

func TestErrSVError(t *testing.T) {
	want := `Test ErrSVError: 1, 'one'`
	if err := newErrSV("Test ErrSVError: %d, '%s'", 1, "one"); err.Error() != want {
		t.Errorf("ErrSV.Error() returned %q, want - %q", err, want)
	}
}

func TestOrigFillFail(t *testing.T) {
	sv := NewStructVerifier(
		func() any { return &struct{B bool}{} },	// creator function
		func(x any) any { return x },				// cloner function
	)

	err := sv.Verify()

	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, because setter for bool was not porvided")
	case errors.As(err, new(*ErrSVOrigFill)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVOrigFill", err, err)
	}
}

func TestRerFillFail(t *testing.T) {
	exhausted := false
	sv := NewStructVerifier(
		func() any { return &struct{B bool}{} },	// creator function
		func(x any) any { return x },				// cloner function
	).AddSetters(func() Setter {
		return func(v reflect.Value) any {
			if exhausted { return nil }
			if _, ok := v.Interface().(bool); ok {
				exhausted = true
				return true
			}
			return nil
		}
	})

	err := sv.Verify()

	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, because reference object should not be filled")
	case errors.As(err, new(*ErrSVRefFill)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVRefFill", err, err)
	}
}

func TestOrigRefEqualFail(t *testing.T) {
	val := false
	sv := NewStructVerifier(
		func() any { return &struct{B bool}{} },	// creator function
		func(x any) any { return x },				// cloner function
	).AddSetters(func() Setter {
		return func(v reflect.Value) any {
			if _, ok := v.Interface().(bool); ok {
				v := val
				val = !val
				return v
			}
			return nil
		}
	})

	err := sv.Verify()

	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, because reference object should not be filled")
	case errors.As(err, new(*ErrSVRefOrigEqual)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVRefOrigEqual", err, err)
	}
}

func Test_autoChangeFail(t *testing.T) {
	sv := NewStructVerifier(
		func() any { return &struct{B bool}{} },	// creator function
		func(x any) any { return x },				// cloner function
	).AddSetters(func() Setter {
		return func(v reflect.Value) any {
			if _, ok := v.Interface().(bool); ok {
				return true
			}
			return nil
		}
	})

	err := sv.Verify()

	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, because changer for bool was not provided")
	case errors.As(err, new(*ErrSVChange)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVChange", err, err)
	}
}

func TestOrigChangedFail(t *testing.T) {
	sv := NewStructVerifier(
		func() any { return &struct{S []int}{} },	// creator function
		func(x any) any { return x },				// cloner function
	).AddSetters(func() Setter {
		return func(v reflect.Value) any {
			if _, ok := v.Interface().([]int); ok {
				return []int{10}
			}
			return nil
		}
	}).AddChangers(func(v reflect.Value) bool {
		is, ok := v.Interface().([]int)
		if !ok {
			return false
		}
		is[0] *= 2

		return true
	})

	err := sv.Verify()
	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, original value should be changed after clone update")
	case errors.As(err, new(*ErrSVOrigChanged)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVOrigChanged", err, err)
	}
}

func TestCloneOrigEqualFail(t *testing.T) {
	sv := NewStructVerifier(
		func() any { return &struct{S []int}{} },	// creator function
		func(x any) any { return x },				// cloner function
	).AddSetters(func() Setter {
		return func(v reflect.Value) any {
			if _, ok := v.Interface().([]int); ok {
				return []int{10}
			}
			return nil
		}
	}).AddChangers(func(v reflect.Value) bool {
		if _, ok := v.Interface().([]int); ok {
			// No not update anything
			return true
		}

		return false
	})

	err := sv.Verify()
	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, clone should be equal original after change")
	case errors.As(err, new(*ErrSVCloneOrigEqual)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVCloneOrigEqual", err, err)
	}
}

func Test_autoChangeFieldNotFound(t *testing.T) {
	sv := NewStructVerifier(
		func() any { return &struct{B bool}{} },	// creator function
		func(x any) any { return x },				// cloner function
	)

	err := sv.autoChange(&struct{B bool}{}, "NxField")

	switch {
	case err == nil:
		t.Errorf("returned no error but must fail, autoChange called for non existing field")
	case errors.As(err, new(*ErrSVFieldNotFound)):
		// OK, expected error
	default:
		t.Errorf("got unexpected error %T (%v), want - *ErrSVFieldNotFound", err, err)
	}
}
