package adapterhttp

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		return strings.Split(f.Tag.Get("json"), ",")[0]
	})
	return v
}
