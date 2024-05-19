package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	*validator.Validate
	ut.Translator
}

func NewValidation(options ...validator.Option) *Validator {
	validate := validator.New(options...)
	return &Validator{Validate: validate}
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
