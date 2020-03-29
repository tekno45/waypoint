package mapper

import (
	"fmt"
	"reflect"
)

// Set is a list of available mapper functions.
type Set []*Func

// Convert converts the input to the output using the set of mappers.
func (s Set) Convert(in, out interface{}) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("output must be a pointer, got %T", out)
	}

	// Dynamically create a function with the correct type
	chain := ChainTarget(CheckReflectType(outVal.Elem().Type()), s, in)
	if chain == nil {
		return fmt.Errorf("no mappers exist to convert %T to %T", in, out)
	}

	// Call the chain
	raw, err := chain.Call()
	if err != nil {
		return err
	}

	// Set the value
	outVal.Elem().Set(reflect.ValueOf(raw))

	return nil
}

// Convert converts the input to the output using the set of mappers.
func (s Set) ConvertSlice(in, out interface{}) error {
	// Get the input slice
	inVal := reflect.ValueOf(in)
	if inVal.Kind() != reflect.Slice {
		return fmt.Errorf("input must be a slice, got %T", in)
	}

	// Get the output slice
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("output must be a pointer, got %T", out)
	}
	outVal = outVal.Elem()
	if outVal.Kind() != reflect.Slice {
		return fmt.Errorf("output pointer value must be a slice, got %T", out)
	}
	outVal.Set(reflect.MakeSlice(outVal.Type(), inVal.Len(), inVal.Len()))

	// Go through each input element
	for i := 0; i < inVal.Len(); i++ {
		from := inVal.Index(i).Interface()
		to := outVal.Index(i).Addr().Interface()
		if err := s.Convert(from, to); err != nil {
			return err
		}
	}

	return nil
}

// ConvertType converts the input to the output type using the set of mappers.
// outType should be a pointer to a nil value of the type you want to convert
// to. Example: (*Foo)(nil). This will return the converted value.
func (s Set) ConvertType(in, outType interface{}) (interface{}, error) {
	// Get the output type and create a value for that type
	typ := reflect.TypeOf(outType).Elem()
	outVal := reflect.New(typ)
	switch typ.Kind() {
	case reflect.Slice:
		if err := s.ConvertSlice(in, outVal.Interface()); err != nil {
			return nil, err
		}

	default:
		if err := s.Convert(in, outVal.Interface()); err != nil {
			return nil, err
		}
	}

	return outVal.Elem().Interface(), nil
}
