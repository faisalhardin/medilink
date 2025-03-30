package model

type MapStaffJourneyPoint struct {
	IDMstStaff        int64 `xorm:"id_mst_staff"`
	IDMstJourneyPoint int64 `xorm:"id_mst_journey_point"`
}

type StaffJoinMstJourneyPoint struct {
	MstStaff        MstStaff          `xorm:"extends"`
	MstJourneyPoint []MstJourneyBoard `xorm:"journey_board"`
}
