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

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return trans, nil
}
