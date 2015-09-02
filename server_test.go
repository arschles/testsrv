package testsrv

import (
	"net/http"
	"testing"
	"time"

	"github.com/arschles/testsrv/Godeps/_workspace/src/github.com/arschles/assert"
)

func TestNoMsgs(t *testing.T) {
	srv := StartServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	recvCh := make(chan []*ReceivedRequest)
	// ensure that it returns after waitTime
	waitTime := 10 * time.Millisecond
	go func() {
		recvCh <- srv.AcceptN(20, waitTime)
	}()
	select {
	case recv := <-recvCh:
		assert.Equal(t, len(recv), 0, "number of received messages")
	case <-time.After(waitTime * 2):
		t.Errorf("AcceptN didn't return after [%+v]", waitTime*2)
	}
}
