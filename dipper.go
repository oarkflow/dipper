package dipper

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/oarkflow/json/sjson"
)

type setOption int

const (
	Zero setOption = 0

	Delete    setOption = 1
	SEPARATOR           = "."
	SLICE               = "#"
)

type Options struct {
	Separator string
	Slice     string
}

type Dipper struct {
	separator string
	slice     string
}

func New(opts Options) *Dipper {
	return &Dipper{separator: opts.Separator, slice: opts.Slice}
}

func groupValues(dataSlice, groupSlice any) (any, error) {
	vA := reflect.ValueOf(groupSlice)
	vB := reflect.ValueOf(dataSlice)
	if vA.Kind() != reflect.Slice || vB.Kind() != reflect.Slice {
		fmt.Println("One or both of the variables are not slices")
		return nil, errors.New("one or both of the variables are not slices")
	}
	if vA.Len() != vB.Len() {
		fmt.Println("Slices have different lengths")
		return nil, errors.New("slices have different lengths")
	}
	resultMap := make(map[string]any)
	for i := 0; i < vA.Len(); i++ {
		keyValue := vA.Index(i).Interface()
		var key string
		switch v := keyValue.(type) {
		case string:
			key = v
		case int:
			key = strconv.Itoa(v)
		default:
			key = fmt.Sprint(v)
		}
		value := vB.Index(i).Interface()
		// If key already exists, append to slice, else set as single value
		if existing, ok := resultMap[key]; ok {
			switch ex := existing.(type) {
			case []any:
				resultMap[key] = append(ex, value)
			default:
				resultMap[key] = []any{ex, value}
			}
		} else {
			resultMap[key] = value
		}
	}
	return resultMap, nil
}

func extract(data any, pattern string) any {
	var results []any
	segments := strings.Split(pattern, ".")
	extractRecursive(data, segments, &results)
	if len(results) == 1 {
		return results[0]
	}
	return results
}

func Extract(data any, pattern string, groupBy string) (any, error) {
	if groupBy == "" {
		return extract(data, pattern), nil
	}
	currentData := extract(data, pattern)
	groupData := extract(data, groupBy)
	return groupValues(currentData, groupData)
}

func extractRecursive(data any, segments []string, results *[]any) {
	if len(segments) == 0 {
		*results = append(*results, data)
		return
	}
	currentSegment := segments[0]
	remainingSegments := segments[1:]
	if m, ok := data.(map[string]any); ok {
		if nextData, exists := m[currentSegment]; exists {
			extractRecursive(nextData, remainingSegments, results)
		}
		return
	}
	if s, ok := data.([]any); ok {
		if currentSegment == "#" {
			for _, item := range s {
				extractRecursive(item, remainingSegments, results)
			}
		} else {
			for _, item := range s {
				extractRecursive(item, segments, results)
			}
		}
		return
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		if currentSegment == "#" {
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				extractRecursive(elem, remainingSegments, results)
			}
		} else {
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				extractRecursive(elem, segments, results)
			}
		}
		return
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(currentSegment)
		if field.IsValid() && field.CanInterface() {
			extractRecursive(field.Interface(), remainingSegments, results)
		}
		return
	}
}

func (d *Dipper) Get(obj any, attribute string, groupBy ...string) (any, error) {
	switch obj := obj.(type) {
	case string:
		// If groupBy is present, extract both and group
		if len(groupBy) > 0 && groupBy[0] != "" {
			val := sjson.Get(obj, attribute)
			group := sjson.Get(obj, groupBy[0])
			if !val.Exists() || !group.Exists() {
				return nil, ErrNotFound
			}
			return groupValues(val.Value(), group.Value())
		}
		rs := sjson.Get(obj, attribute)
		if !rs.Exists() {
			return nil, ErrNotFound
		}
		return rs.Value(), nil
	case []byte:
		if len(groupBy) > 0 && groupBy[0] != "" {
			val := sjson.GetBytes(obj, attribute)
			group := sjson.GetBytes(obj, groupBy[0])
			if !val.Exists() || !group.Exists() {
				return nil, ErrNotFound
			}
			return groupValues(val.Value(), group.Value())
		}
		rs := sjson.GetBytes(obj, attribute)
		if !rs.Exists() {
			return nil, ErrNotFound
		}
		return rs.Value(), nil
	default:
		return d.get(obj, attribute, groupBy...)
	}
}

func (d *Dipper) get(obj any, attribute string, groupBy ...string) (any, error) {
	if strings.Contains(attribute, SLICE) {
		if len(groupBy) > 0 {
			return Extract(obj, attribute, groupBy[0])
		}
		return Extract(obj, attribute, "")
	}
	val, _, err := GetReflectValue(reflect.ValueOf(obj), attribute, d.separator, false)
	if err != nil {
		return nil, err
	}
	return val.Interface(), nil
}

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
			keyKind := value.Type().Key().Kind()
			if keyKind != reflect.String && keyKind != reflect.Interface {
				return value, "", ErrMapKeyNotString
			}
			if toSet && i == maxSetDepth {
				return value, fieldName, nil
			}
			mapValue := value.MapIndex(reflect.ValueOf(fieldName))
			if !mapValue.IsValid() {
				return value, "", ErrNotFound
			}
			value = mapValue
		case reflect.Struct:
			field, ok := value.Type().FieldByName(fieldName)
			if !ok {
				return value, "", ErrNotFound
			}
			if field.PkgPath != "" {
				return value, "", ErrUnexported
			}
			value = value.FieldByName(fieldName)
		case reflect.Slice, reflect.Array:
			if i == 0 && fieldName == "" {
				break
			}
			if strings.HasPrefix(fieldName, "[") && strings.HasSuffix(fieldName, "]") {
				fieldName = fieldName[1 : len(fieldName)-1]
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
