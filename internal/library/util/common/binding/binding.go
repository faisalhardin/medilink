package binding

import (
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"

	"github.com/pkg/errors"

	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/util/validation"
	customtime "github.com/faisalhardin/medilink/pkg/type/time"
	"github.com/go-playground/locales/en"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/gorilla/schema"
)

var (
	validatorURL  *validation.Validator
	validatorJSON *validation.Validator
	schemaDecoder = schema.NewDecoder()
	json          = jsoniter.ConfigCompatibleWithStandardLibrary

	ErrInvalidContentType = errors.New("unrecognized content type")
)

const (
	ContentURLEncoded string = "application/x-www-form-urlencoded"
	ContentJSON       string = "application/json"
	ContentFormData   string = "multipart/form-data"
	ContentType       string = "Content-Type"
)

func init() {

	schemaDecoder.RegisterConverter(customtime.Time{}, customtime.TimeConverter)

	validatorJSON = validation.NewValidation()
	validatorURL = validation.NewValidation()

	englishTranslator := validation.NewTranslator(en.New())

	validatorJSON.SetTranslator(englishTranslator)
	validatorJSON.RegisterTagNameFunc(validation.RegisterJSONTagFunc)
	validatorJSON.TranslateOverride(
		validation.SetCustomRequiredMessage(),
		validation.SetCustomEmailMessage(),
		validation.SetCustomMaxNumberCharacterMessage(),
	)

	validatorURL.RegisterTagNameFunc(validation.RegisterSchemaTag)
	validatorURL.SetTranslator(englishTranslator)
	validatorURL.TranslateOverride(
		validation.SetCustomRequiredMessage(),
		validation.SetCustomEmailMessage(),
		validation.SetCustomMaxNumberCharacterMessage(),
	)

	schemaDecoder.RegisterConverter(model.Time{}, model.TimeConverter)
	schemaDecoder.IgnoreUnknownKeys(true)

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
		if err = validatorJSON.Struct(targetDecode); err != nil {
			return commonerr.NewErrorMessage().SetTranslator(validatorJSON).SetBadRequest().SetErrorValidator(err)
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

	if err := schemaDecoder.Decode(targetDecode, value); err != nil {
		return err
	}
	if err := validatorURL.Struct(targetDecode); err != nil {
		return commonerr.NewErrorMessage().SetTranslator(validatorURL).SetBadRequest().SetErrorValidator(err)
	}

	return nil
}
