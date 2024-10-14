package commonerr

import "errors"

func SetNewNoInstitutionError() error {
	return errors.New("no institution for request")
}

func SetNoVisitDetailError() error {
	return SetNewBadRequest("invalid", "no patient visit detail found")
}
