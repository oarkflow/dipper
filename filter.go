package dipper

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var filterRegex = regexp.MustCompile(`(?m)^([\w-]*)==?(.*)$`)

func filterSlice(value reflect.Value, fieldName string) (reflect.Value, error) {
	if !strings.Contains(fieldName, "=") {
		return reflect.Value{}, nil
	}
	match := filterRegex.FindStringSubmatch(fieldName)
	if match == nil {
		return reflect.Value{}, ErrInvalidFilterExpression
	}
	parseFilterValue := func(v string) (interface{}, error) {
		if strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'") {
			return v[1 : len(v)-1], nil
		}
		if v == "true" || v == "false" {
			return v == "true", nil
		}
		if v == "null" {
			return nil, nil
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return parsed, nil
		}
		return v, nil // fallback to string
	}
	filterKey := match[1]
	filterValue, err := parseFilterValue(match[2])
	if err != nil {
		return reflect.Value{}, err
	}
	compareValues := func(v reflect.Value) bool {
		v = getElemSafe(v)
		if fv, ok := filterValue.(float64); ok {
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return float64(v.Int()) == fv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return float64(v.Uint()) == fv
			case reflect.Float32, reflect.Float64:
				return v.Float() == fv
			}
		}
		return reflect.DeepEqual(v.Interface(), filterValue)
	}

	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		itemSafe := getElemSafe(item)
		switch itemSafe.Kind() {
		case reflect.Map:
			for _, mapKey := range itemSafe.MapKeys() {
				if mapKey.String() != filterKey {
					continue
				}
				if compareValues(itemSafe.MapIndex(mapKey)) {
					return item, nil
				}
			}
		case reflect.Struct:
			for j := 0; j < itemSafe.NumField(); j++ {
				if itemSafe.Type().Field(j).Name != filterKey {
					continue
				}
				field := itemSafe.Field(j)
				if compareValues(field) {
					return item, nil
				}
			}
		default:
			if filterKey == "" && compareValues(item) {
				return item, nil
			}
		}
	}
	return reflect.Value{}, ErrFilterNotFound
}
