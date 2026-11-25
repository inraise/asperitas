package v1

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/inraise/asperitas/internal/domain"
)

func newValidator() *validator.Validate {
	v := validator.New()

	// В ошибках показываем имя из json-тега ("username"), а не "Username".
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	must(v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return domain.ValidUsername(fl.Field().String())
	}))
	must(v.RegisterValidation("maxbytes", func(fl validator.FieldLevel) bool {
		n, err := strconv.Atoi(fl.Param())
		return err == nil && len(fl.Field().String()) <= n
	}))
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// validateStruct возвращает ошибки полей в формате контракта (nil — всё ок).
func validateStruct(v *validator.Validate, s any) []fieldError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return []fieldError{{Location: "body", Param: "body", Msg: "is invalid"}}
	}
	out := make([]fieldError, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, fieldError{Location: "body", Param: fe.Field(), Msg: message(fe)})
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + fe.Param() + " characters long"
	case "max":
		return "must be at most " + fe.Param() + " characters long"
	case "maxbytes":
		return "must be at most " + fe.Param() + " bytes long"
	case "username":
		return "may contain only letters, digits, '_' and '-' (up to 32 characters)"
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "is invalid"
	}
}
