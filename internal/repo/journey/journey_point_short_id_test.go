package journey

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

func openJourneyTestDB(t *testing.T) *JourneyDB {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "postgresql://postgres@127.0.0.1:5432/postgres?sslmode=disable&timezone=UTC"
	}
	eng, err := xorm.NewEngine("postgres", dsn)
	if err != nil {
		t.Skipf("postgres engine: %v", err)
	}
	if err := eng.Ping(); err != nil {
		t.Skipf("postgres ping: %v", err)
	}
	return &JourneyDB{DB: &xormlib.DBConnect{MasterDB: eng, SlaveDB: eng}}
}

func assertColumnNotFound(t *testing.T, label string, got *model.MstJourneyPoint, err error) {
	t.Helper()
	if !errors.Is(err, constant.ErrorRowNotFound) {
		t.Fatalf("%s: err=%v got=%+v, want ErrorRowNotFound", label, err, got)
	}
}

func TestGetJourneyPointByShortID_ClinicScopedLiveOnly(t *testing.T) {
	db := openJourneyTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	suffix := fmt.Sprintf("%d", now.UnixNano()%1_000_000_000)
	shortA := ("a" + suffix)[:8]
	shortB := ("b" + suffix)[:8]

	instA := model.Institution{Name: "be7-scope-a-" + suffix, StaffNumber: 1, MaxStaff: 10}
	instB := model.Institution{Name: "be7-scope-b-" + suffix, StaffNumber: 1, MaxStaff: 10}
	if _, err := db.DB.MasterDB.Table("mdl_mst_institution").Insert(&instA); err != nil {
		t.Fatalf("insert clinic A: %v", err)
	}
	if _, err := db.DB.MasterDB.Table("mdl_mst_institution").Insert(&instB); err != nil {
		t.Fatalf("insert clinic B: %v", err)
	}
	boardA := model.MstJourneyBoard{Name: "board-a", IDMstInstitution: instA.ID}
	boardB := model.MstJourneyBoard{Name: "board-b", IDMstInstitution: instB.ID}
	if _, err := db.DB.MasterDB.Table("mdl_mst_journey_board").Insert(&boardA); err != nil {
		t.Fatalf("insert board A: %v", err)
	}
	if _, err := db.DB.MasterDB.Table("mdl_mst_journey_board").Insert(&boardB); err != nil {
		t.Fatalf("insert board B: %v", err)
	}
	pointA := model.MstJourneyPoint{
		Name: "col-a", ShortID: shortA, Position: 1,
		IDMstJourneyBoard: boardA.ID, IDMstInstitution: instA.ID,
	}
	pointB := model.MstJourneyPoint{
		Name: "col-b", ShortID: shortB, Position: 1,
		IDMstJourneyBoard: boardB.ID, IDMstInstitution: instB.ID,
	}
	if _, err := db.DB.MasterDB.Table("mdl_mst_journey_point").Insert(&pointA); err != nil {
		t.Fatalf("insert point A: %v", err)
	}
	if _, err := db.DB.MasterDB.Table("mdl_mst_journey_point").Insert(&pointB); err != nil {
		t.Fatalf("insert point B: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.DB.MasterDB.Exec("DELETE FROM mdl_mst_journey_point WHERE id IN (?, ?)", pointA.ID, pointB.ID)
		_, _ = db.DB.MasterDB.Exec("DELETE FROM mdl_mst_journey_board WHERE id IN (?, ?)", boardA.ID, boardB.ID)
		_, _ = db.DB.MasterDB.Exec("DELETE FROM mdl_mst_institution WHERE id IN (?, ?)", instA.ID, instB.ID)
	})

	got, err := db.GetJourneyPointByShortID(ctx, shortB, instA.ID)
	assertColumnNotFound(t, "clinic B short_id via clinic A", got, err)

	got, err = db.GetJourneyPointByShortID(ctx, shortA, instA.ID)
	if err != nil {
		t.Fatalf("own live column: %v", err)
	}
	if got == nil || got.ID != pointA.ID {
		t.Fatalf("own live column got=%+v want id=%d", got, pointA.ID)
	}

	archived := now
	pointA.DeleteTime = &archived
	if _, err := db.DB.MasterDB.Table("mdl_mst_journey_point").ID(pointA.ID).Unscoped().Cols("delete_time").Update(&pointA); err != nil {
		t.Fatalf("archive point A: %v", err)
	}
	got, err = db.GetJourneyPointByShortID(ctx, shortA, instA.ID)
	assertColumnNotFound(t, "archived column", got, err)

	got, err = db.GetJourneyPointByShortID(ctx, "zzzzmiss", instA.ID)
	assertColumnNotFound(t, "missing short_id", got, err)
}
