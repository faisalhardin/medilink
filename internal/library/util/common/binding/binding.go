package binding

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/faisalhardin/auth-vessel/internal/library/common/commonerr"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gorilla/schema"
)

var (
	validatorURL  *validator.Validate
	validatorJSON *validator.Validate

	validatorTranslator *ut.Translator

	ErrInvalidContentType = errors.New("unrecognized content type")
)

const (
	ContentURLEncoded string = "application/x-www-form-urlencoded"
	ContentJSON       string = "application/json"
	ContentFormData   string = "multipart/form-data"
	ContentType       string = "Content-Type"
)

func init() {
	validatorJSON = validator.New()
	validatorURL = validator.New()

	english := en.New()
	translatorLib := ut.New(english)
	translator, _ := translatorLib.GetTranslator("en")
	validatorTranslator = &translator

	entranslations.RegisterDefaultTranslations(validatorJSON, translator)

	validatorJSON.RegisterTagNameFunc(registerJSONTagfunc)
	validatorURL.RegisterTagNameFunc(registerSchemaTag)

	translateOverride(translator, validatorJSON)

}

func translateOverride(trans ut.Translator, v *validator.Validate) {

	v.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "The {0} field must have a value", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())

		return t
	})

	v.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "The {0} field must be no longer than {1} characters", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		param := fe.Param()
		tag := fe.Tag()
		t, _ := ut.T(tag, fe.Field(), param)

		return t
	})

	v.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "The {0} field is invalid. Please double check and correct the data.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())

		return t
	})

}

func registerSchemaTag(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("schema"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func registerJSONTagfunc(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func Bind(r *http.Request, targetDecode interface{}) error {
	if r.Method == http.MethodGet {
		if err := decodeSchemaRequest(r, targetDecode); err != nil {
			return err
		}
		return nil
	}

	contentType := filterFlags(r.Header.Get(ContentType))

	switch contentType {
	case ContentURLEncoded:
		err := r.ParseForm()
		if err != nil {
			return err
		}
		if err := decodeSchemaRequest(r, targetDecode); err != nil {
			return err
		}
	case ContentJSON:
		bodyDecode := json.NewDecoder(r.Body)
		err := bodyDecode.Decode(targetDecode)
		if err != nil {
			return commonerr.SetNewBadRequest("invalid body", err.Error())
		}
		if err := validatorJSON.Struct(targetDecode); err != nil {
			return commonerr.NewErrorMessage().SetTranslator(*validatorTranslator).SetBadRequest().SetErrorValidator(err)
		}
	case ContentFormData:
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			return err
		}
		if err := decodeSchemaRequest(r, targetDecode); err != nil {
			return err
		}
	default:
		return ErrInvalidContentType
	}

	return nil

}

func decodeSchemaRequest(r *http.Request, val interface{}) error {
	sourceDecode := r.Form
	if r.Method == http.MethodGet {
		sourceDecode = r.URL.Query()
	}
	return BindQuery(sourceDecode, val)
}

func BindQuery(value url.Values, targetDecode interface{}) error {
	decoder := schema.NewDecoder()

	if err := decoder.Decode(targetDecode, value); err != nil {
		return err
	}
	if err := validatorURL.Struct(targetDecode); err != nil {
		return commonerr.NewErrorMessage().SetBadRequest().SetErrorValidator(err)
	}

	return nil
}
