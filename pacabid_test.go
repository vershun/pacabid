package main

import "testing"

func TestBlah(t *testing.T) {
	pb := newPacabid()

	go func() {
		if err := pb.loopBarFeed(); err != nil {
			t.Fatal(err)
		}
	}()
	select {
	case hi := <-pb.marketStream:
		t.Errorf("%+v", hi)
	}
}
