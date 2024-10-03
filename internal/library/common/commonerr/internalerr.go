package commonerr

import "errors"

func SetNewNoInstitutionError() error {
	return errors.New("no institution for request")
}
