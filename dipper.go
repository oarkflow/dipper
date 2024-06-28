package dipper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/oarkflow/dipper/sjson"
)

// setOption is a type used for special assignments in a set operation.
type setOption int

const (
	// Zero is used as the new value in Set() to set the attribute to its zero
	// value (e.g. "" for string, nil for any, etc.).
	Zero setOption = 0
	// Delete is used as the new value in Set() to delete a map key. If the
	// field is not a map value, the value will be zeroed (see Zero).
	Delete          setOption = 1
	SEPARATOR                 = "."
	SLICE                     = "#"
	SEPARATOR_SLICE           = ".#"
)

// Options defines the configuration of a Dipper instance.
type Options struct {
	Separator string
	Slice     string
}

// Dipper allows to access deeply-nested object attributes to get or set their
// values. Attributes are specified by a string with its fields separated by
// some delimiter (e.g. “Books.3.Author" or "Books->3->Author", with "." and
// "->" as delimiters, respectively).
type Dipper struct {
	separator string
	slice     string
}

// New returns a new Dipper instance.
func New(opts Options) *Dipper {
	return &Dipper{separator: opts.Separator, slice: opts.Slice}
}

// Get returns the value of the given obj attribute. The attribute uses some
// delimiter-notation to allow accessing nested fields, slice elements or map
// keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// If an error occurs, it will be returned as the attribute value, so it should
// be handled. All the returned errors are fieldError.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.Get(myObj, "SomeStructField.1.some_key_map")
//		if err := Error(v); err != nil {
//		    return err
//		}
func (d *Dipper) Get(obj any, attribute string) (any, error) {
	switch obj := obj.(type) {
	case string:
		rs := sjson.Get(obj, attribute)
		if !rs.Exists() {
			return nil, ErrNotFound
		}
		return rs.Value(), nil
	case []byte:
		rs := sjson.GetBytes(obj, attribute)
		if !rs.Exists() {
			return nil, ErrNotFound
		}
		return rs.Value(), nil
	default:
		return d.get(reflect.ValueOf(obj), attribute)
	}
}

func (d *Dipper) get(value reflect.Value, attribute string) (any, error) {
	if strings.HasPrefix(attribute, SLICE) {
		attribute = SEPARATOR + attribute
	}
	if strings.Contains(attribute, d.separator+SLICE+d.separator) {
		// Handle nested slices
		fields := strings.Split(attribute, d.separator+SLICE+d.separator)
		left := fields[0]
		right := fields[1]
		val, _, err := GetReflectValue(value, left, d.separator, false)
		if err != nil {
			return nil, err
		}
		return d.getSliceValues(val, right)
	} else if strings.HasPrefix(attribute, SEPARATOR_SLICE) {
		// Handle top-level slice with a dot prefix
		return d.getSliceValues(value, attribute[3:])
	} else {
		// Handle single field access
		val, _, err := GetReflectValue(value, attribute, d.separator, false)
		if err != nil {
			return nil, err
		}
		return val.Interface(), nil
	}
}

func (d *Dipper) getSliceValues(value reflect.Value, attribute string) (any, error) {
	var results []any
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			elem := value.Index(i)
			if attribute == "" || attribute == SLICE {
				results = append(results, elem.Interface())
			} else {
				subVal, _, err := GetReflectValue(elem, attribute, d.separator, false)
				if err != nil {
					return nil, err
				}
				results = append(results, subVal.Interface())
			}
		}
	default:
		results = append(results, value.Interface())
	}

	return results, nil
}

// GetMany returns a map with the values of the given obj attributes.
// It works as Dipper.Get(), but it takes a slice of attributes to return their
// corresponding values. The returned map will have the same length as the
// attributes slice, with the attributes as keys.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.GetMany(myObj, []string{"Name", "Age", "Skills.skydiving})
//		if err := v.FirstError(); err != nil {
//		    return err
//		}
func (d *Dipper) GetMany(obj any, attributes []string) (Fields, error) {
	m := make(Fields, len(attributes))

	for _, attr := range attributes {
		if _, ok := m[attr]; !ok {
			t, err := d.Get(obj, attr)
			if err != nil {
				return nil, err
			}
			m[attr] = t
		}
	}

	return m, nil
}

// Set sets the value of the given obj attribute to the new provided value.
// The attribute uses some delimiter-notation to allow accessing nested fields,
// slice elements or map keys. Field names and key maps are case-sensitive.
// All the struct fields accessed must be exported.
// ErrUnaddressable will be returned if obj is not addressable.
// It returns nil if the value was successfully set, otherwise it will return
// a fieldError.
//
// Example:
//
//	 // Using "." as the Dipper separator
//		v := my_dipper.Set(&myObj, "SomeStructField.1.some_key_map", 123)
//		if err != nil {
//		    return err
//		}
func (d *Dipper) Set(obj any, attribute string, new any) error {
	var err error

	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var lastField string
	value, lastField, err = GetReflectValue(value, attribute, d.separator, true)
	if err != nil {
		return err
	}

	var optZero, optDelete bool

	var newValue reflect.Value
	switch new {
	case Zero:
		optZero = true
	case Delete:
		optDelete = true
	default:
		newValue = reflect.ValueOf(new)
		if newValue.Kind() == reflect.Ptr {
			newValue = newValue.Elem()
		}
	}

	if value.Kind() == reflect.Map {
		if !optZero && !optDelete {
			mapValueType := value.Type().Elem()
			if mapValueType.Kind() != reflect.Interface && mapValueType != newValue.Type() {
				return ErrTypesDoNotMatch
			}
		}

		// Initialize map if needed
		if value.IsNil() {
			keyType := value.Type().Key()
			valueType := value.Type().Elem()
			mapType := reflect.MapOf(keyType, valueType)
			value.Set(reflect.MakeMapWithSize(mapType, 0))
		}

		value.SetMapIndex(reflect.ValueOf(lastField), newValue)
	} else {
		if !optZero && !optDelete {
			if !value.CanAddr() {
				return ErrUnaddressable
			}
			if value.Kind() != reflect.Interface && value.Type() != newValue.Type() {
				return ErrTypesDoNotMatch
			}
		} else {
			newValue = reflect.Zero(value.Type())
		}
		value.Set(newValue)
	}
	return nil
}

// GetReflectValue gets the reflect.Value of the given value attribute.
// It splits the attribute into the field names, map keys and slice indexes
// and uses reflection to get the final value.
// toSet indicates that the function must return a value that will be set to
// another value, which is used in the special case of maps (maps elements are
// not addressable).
// It also returns the name of the accessed field.
func GetReflectValue(value reflect.Value, attribute string, sep string, toSet bool) (_ reflect.Value, fieldName string, _ error) {
	if attribute == "" {
		return value, "", nil
	}

	if len(sep) == 0 {
		sep = "."
	}

	splitter := newAttributeSplitter(attribute, sep)

	var i, maxSetDepth int
	if toSet {
		maxSetDepth = splitter.CountRemaining() - 1
	}

	for splitter.HasMore() {
		fieldName, i = splitter.Next()
		value = getElemSafe(value)

		switch value.Kind() {
		case reflect.Map:
			// Check that the map accept string keys
			keyKind := value.Type().Key().Kind()
			if keyKind != reflect.String && keyKind != reflect.Interface {
				return value, "", ErrMapKeyNotString
			}

			// If a map key has to be set, skip the last attribute and return the map
			if toSet && i == maxSetDepth {
				return value, fieldName, nil
			}

			mapValue := value.MapIndex(reflect.ValueOf(fieldName))
			if !mapValue.IsValid() {
				fmt.Println(fieldName)
				return value, "", ErrNotFound
			}

			value = mapValue

		case reflect.Struct:
			field, ok := value.Type().FieldByName(fieldName)
			if !ok {
				fmt.Println(fieldName)
				return value, "", ErrNotFound
			}
			// Check if field is unexported (method IsExported() was introduced in Go 1.17)
			if field.PkgPath != "" {
				return value, "", ErrUnexported
			}

			value = value.FieldByName(fieldName)

		case reflect.Slice, reflect.Array:
			// Ignores field if it is the first one and it is empty. This
			// happens when using brackets on a root slice (e.g. "[1].Name").
			if i == 0 && fieldName == "" {
				break
			}

			if strings.HasPrefix(fieldName, "[") && strings.HasSuffix(fieldName, "]") {
				fieldName = fieldName[1 : len(fieldName)-1]

				// Try to apply the filter to the slice elements
				foundValue, err := filterSlice(value, fieldName)
				if err != nil {
					return value, "", err
				}
				if foundValue.IsValid() {
					value = foundValue
					break
				}
			}

			sliceIndex, err := strconv.Atoi(fieldName)
			if err != nil {
				return value, "", ErrInvalidIndex
			}
			if sliceIndex < 0 || sliceIndex >= value.Len() {
				return value, "", ErrIndexOutOfRange
			}
			field := value.Index(sliceIndex)
			value = field

		default:
			return value, "", ErrNotFound
		}
	}

	return value, fieldName, nil
}

// getElemSafe returns the underlying value of an interface/pointer reflect.Value.
func getElemSafe(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Func, reflect.Chan, reflect.Map, reflect.Slice:
		if v.IsNil() {
			return v
		}
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}
