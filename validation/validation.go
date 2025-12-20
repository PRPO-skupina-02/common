package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

func GetDefaultValidationEngine() (*validator.Validate, error) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil, errors.New("Failed to get default validation engine")
	}
	return v, nil
}

func RegisterValidation() (ut.Translator, error) {

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	v, err := GetDefaultValidationEngine()
	if err != nil {
		return nil, err
	}

	err = en_translations.RegisterDefaultTranslations(v, trans)
	if err != nil {
		return nil, err
	}

	err = v.RegisterValidation("non-nil-uuid", nonNilUUID)
	if err != nil {
		return nil, err
	}
	err = v.RegisterTranslation("non-nil-uuid", trans, func(ut ut.Translator) error {
		return ut.Add("non-nil-uuid", "{0} must not be a nil uuid!", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("non-nil-uuid", fe.Field())

		return t
	})
	if err != nil {
		return nil, err
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return trans, nil
}

var nonNilUUID validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	uuidValue, err := uuid.Parse(value)
	if err != nil {
		return false
	}
	return uuidValue != uuid.Nil
}
