package journey

import (
	"context"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/lib/pq"
	"xorm.io/builder"
)

func TestListServicePointsEmptyWithoutClinic(t *testing.T) {
	db := &JourneyDB{} // nil DB; must not be touched
	for _, clinic := range []int64{0, -1} {
		rows, err := db.ListServicePoints(context.Background(), model.GetServicePointParams{
			IDMstInstitution: clinic,
			ID:               []int64{1, 2},
			IDMstBoard:       9,
		})
		if err != nil {
			t.Fatalf("clinic %d: %v", clinic, err)
		}
		if rows == nil || len(rows) != 0 {
			t.Fatalf("clinic %d: got %#v, want empty slice", clinic, rows)
		}
	}
}

func TestServicePointListCondBindsIDAndClinic(t *testing.T) {
	sql, args, err := builder.ToSQL(servicePointListCond(model.GetServicePointParams{
		IDMstInstitution: 7,
		ID:               []int64{3, 5},
		Name:             []string{"poli"},
		IDMstBoard:       4,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "id_mst_institution=? AND (id = ANY(?)) AND (name ilike ANY(?)) AND id_mst_journey_board=?"; sql != want {
		t.Fatalf("sql=%q want %q", sql, want)
	}
	if len(args) != 4 {
		t.Fatalf("args=%d %#v", len(args), args)
	}
	if args[0] != int64(7) {
		t.Fatalf("clinic arg=%v", args[0])
	}
	idArr, ok := args[1].(*pq.Int64Array)
	if !ok {
		t.Fatalf("id arg type %T", args[1])
	}
	if len(*idArr) != 2 || (*idArr)[0] != 3 || (*idArr)[1] != 5 {
		t.Fatalf("id arr=%v", *idArr)
	}
	if args[3] != int64(4) {
		t.Fatalf("board arg=%v", args[3])
	}
}

func TestServicePointListCondOmitsBoardWhenZero(t *testing.T) {
	sql, args, err := builder.ToSQL(servicePointListCond(model.GetServicePointParams{
		IDMstInstitution: 7,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if sql != "id_mst_institution=?" || len(args) != 1 || args[0] != int64(7) {
		t.Fatalf("sql=%q args=%#v", sql, args)
	}
}
