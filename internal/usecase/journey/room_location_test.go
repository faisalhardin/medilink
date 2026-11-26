package journey

import (
	"context"
	"testing"
)

func TestNoopRoomLocation(t *testing.T) {
	var port RoomLocationPort = NoopRoomLocation{}
	got, err := port.Read(context.Background(), 7, []int64{1})
	if err != nil || got != nil {
		t.Fatalf("Read: got %#v err %v", got, err)
	}
	if err := port.Enqueue(context.Background(), 7, 1, "create"); err != nil {
		t.Fatal(err)
	}
	locked, err := port.Clearing(context.Background(), 7, 0)
	if err != nil || locked {
		t.Fatalf("Clearing locked=%v err=%v", locked, err)
	}
}

func TestNilRoomLocationPortActsAsNoop(t *testing.T) {
	got, err := roomLocationPort(nil).Read(context.Background(), 7, []int64{1})
	if err != nil || got != nil {
		t.Fatalf("got %#v err %v", got, err)
	}
}
