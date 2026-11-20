package permission

import (
	"context"
	"testing"
)

func TestHasStaffPermissionWithoutStaffOrClinicIsDeniedBeforeQuery(t *testing.T) {
	c := &Conn{} // nil DB: reaching the query would panic
	for _, tt := range []struct{ staff, clinic int64 }{{0, 7}, {4, 0}, {4, -1}} {
		got, err := c.HasStaffPermission(context.Background(), tt.staff, tt.clinic, "service_point.manage")
		if got || err != nil {
			t.Fatalf("staff %d clinic %d: got %v, %v; want false, nil", tt.staff, tt.clinic, got, err)
		}
	}
}
