package journey

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/pkg/errors"
)

// The four names are one contract. At this boundary every hidden room is
// already ErrorNoAffectedRow, so one fake covers missing, other clinic, orphan,
// and archived.
func TestRoomLookupIs404(t *testing.T) {
	for _, tc := range []struct{ name string }{
		{"missing"}, {"other_clinic"}, {"orphan"}, {"archived"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uc := &JourneyUC{JourneyDB: hiddenRoomDB{}}
			ctx := context.Background()
			for verb, call := range map[string]func() error{
				"get": func() error {
					_, err := uc.GetServicePoint(ctx, 99)
					return err
				},
				"update": func() error {
					return uc.UpdateServicePoint(ctx, &model.MstServicePoint{ID: 99, Name: "x"})
				},
				"archive": func() error {
					return uc.ArchiveServicePoint(ctx, &model.MstServicePoint{ID: 99})
				},
			} {
				assertSet404Wire(t, verb, call())
			}
		})
	}
}

// hiddenRoomDB answers every room lookup the way UserJourneyDB does when the
// row is not visible to the caller's clinic. Any other call panics.
type hiddenRoomDB struct{ journeyRepo.JourneyDB }

func (hiddenRoomDB) GetServicePoint(context.Context, model.MstServicePoint) (*model.MstServicePoint, error) {
	return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
}

func (hiddenRoomDB) UpdateServicePoint(context.Context, *model.MstServicePoint) error {
	return errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
}

func (hiddenRoomDB) DeleteServicePoint(context.Context, *model.MstServicePoint) error {
	return errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
}

func assertSet404Wire(t *testing.T, verb string, err error) {
	t.Helper()
	got := httptest.NewRecorder()
	want := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
		t.Fatalf("%s: SetError: %v", verb, setErr)
	}
	if setErr := commonwriter.SetError(context.Background(), want, commonerr.Set404()); setErr != nil {
		t.Fatalf("%s: SetError want: %v", verb, setErr)
	}
	if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("%s: got %d %s want %d %s", verb, got.Code, got.Body.Bytes(), want.Code, want.Body.Bytes())
	}
}
