package dipper

type fieldError string

func (e fieldError) Error() string {
	return string(e)
}

const (
	ErrNotFound                = fieldError("dipper: field not found")
	ErrInvalidIndex            = fieldError("dipper: invalid index")
	ErrIndexOutOfRange         = fieldError("dipper: index out of range")
	ErrMapKeyNotString         = fieldError("dipper: map key is not of string type")
	ErrUnexported              = fieldError("dipper: field is unexported")
	ErrUnaddressable           = fieldError("dipper: field is unaddressable")
	ErrTypesDoNotMatch         = fieldError("dipper: value type does not match field type")
	ErrInvalidFilterExpression = fieldError("dipper: invalid search expression")
	ErrFilterNotFound          = fieldError("dipper: no matches for filter expression")
	ErrInvalidFilterValue      = fieldError("dipper: invalid value for filter expression")
)

func IsFieldError(v any) bool {
	_, ok := v.(fieldError)
	return ok
}

func Error(v any) error {
	if err, ok := v.(fieldError); ok {
		return err
	}
	return nil
}

func (f Fields) HasErrors() bool {
	return f.FirstError() != nil
}

func (f Fields) FirstError() error {
	for _, v := range f {
		if IsFieldError(v) {
			return v.(fieldError)
		}
	}
	return nil
}
