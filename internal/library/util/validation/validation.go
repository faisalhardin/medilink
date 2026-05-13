package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/volatiletech/null/v8"
)

type Validator struct {
	*validator.Validate
	ut.Translator
}

func NewValidation(options ...validator.Option) *Validator {
	validate := validator.New(options...)
	registerNullV8Types(validate)
	return &Validator{Validate: validate}
}

// registerNullV8Types maps github.com/volatiletech/null/v8 types to plain values so
// tags like max, min, lte apply without "Bad field type null.String" panics.
func registerNullV8Types(v *validator.Validate) {
	fn := func(field reflect.Value) interface{} {
		switch val := field.Interface().(type) {
		case null.String:
			if !val.Valid {
				return ""
			}
			return val.String
		case null.Int16:
			if !val.Valid {
				return int64(0)
			}
			return int64(val.Int16)
		case null.Int64:
			if !val.Valid {
				return int64(0)
			}
			return val.Int64
		case null.Float32:
			if !val.Valid {
				return float64(0)
			}
			return float64(val.Float32)
		case null.Bool:
			if !val.Valid {
				return false
			}
			return val.Bool
		default:
			return nil
		}
	}
	// One registration per concrete type (validator API).
	v.RegisterCustomTypeFunc(fn, null.String{})
	v.RegisterCustomTypeFunc(fn, null.Int16{})
	v.RegisterCustomTypeFunc(fn, null.Int64{})
	v.RegisterCustomTypeFunc(fn, null.Float32{})
	v.RegisterCustomTypeFunc(fn, null.Bool{})
}

func (v *Validator) SetTranslator(translator ut.Translator) {
	v.Translator = translator
	entranslations.RegisterDefaultTranslations(v.Validate, v.Translator)
}

type RegisterTranslationHandler func(v *Validator) (err error)

func (v *Validator) TranslateOverride(translationRegistrationaHandler ...RegisterTranslationHandler) {

	if v.Translator == nil {
		return
	}

	for _, f := range translationRegistrationaHandler {
		f(v)
	}

}

func SetCustomValidationMessage(tag string, errorValidationMessage string) RegisterTranslationHandler {
	return func(v *Validator) (err error) {
		err = v.RegisterTranslation(tag, v, func(ut ut.Translator) error {
			return ut.Add(tag, errorValidationMessage, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			param := fe.Param()
			tag := fe.Tag()
			t, _ := ut.T(tag, fe.Field(), param)

			return t
		})
		return err
	}
}

func SetCustomRequiredMessage() RegisterTranslationHandler {
	return SetCustomValidationMessage("required", "The {0} field must have a value.")
}

func SetCustomMaxNumberCharacterMessage() RegisterTranslationHandler {
	return SetCustomValidationMessage("max", "The {0} field must be no longer than {1} characters.")
}

func SetCustomEmailMessage() RegisterTranslationHandler {
	return SetCustomValidationMessage("email", "The {0} field format is invalid.")
}

func NewTranslator(englishTranslator locales.Translator) ut.Translator {
	translatorLib := ut.New(englishTranslator)
	translator, _ := translatorLib.GetTranslator("en")
	return translator
}

func RegisterSchemaTag(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("schema"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func RegisterJSONTagFunc(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}
