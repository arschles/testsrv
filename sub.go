package httpsub

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/arschles/httpsub/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
)

const (
	UUIDHeaderName = "HTTP-Sub-UUID"
)

// ReceivedRequest represents a request that the subscriber received.
type ReceivedRequest struct {
	// The request the subscriber received
	Request *http.Request
	// The receipt time
	Time time.Time
	// Unique identifier of the request. will be returned as UUIDHeaderName in each
	// response that the subscriber sends
	UUID string
}

// Subscriber is a HTTP server that you can send real HTTP requests to
// in your unit tests. In addition to the functionality of net/http/httptest,
// it allows callers to easily retrieve the requests that came into the server
type Subscriber struct {
	ch  chan *ReceivedRequest
	srv *httptest.Server
}

func StartSubscriber(handler http.Handler) *Subscriber {
	ch := make(chan *ReceivedRequest)
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := uuid.New()
		go func() {
			// TODO: stop sending on s.Close()
			recvReq := ReceivedRequest{Request: r, Time: time.Now(), UUID: u}
			ch <- &recvReq
		}()
		w.Header().Add("HTTP-Sub-UUID", u)
		handler.ServeHTTP(w, r)
	})
	s := httptest.NewServer(wrappedHandler)
	return &Subscriber{ch: ch, srv: s}
}

// Close releases all resources on this subscriber
func (s *Subscriber) Close() {
	s.srv.Close()
}

// URLStr returns the string representation of the URL to call to hit this server
func (s *Subscriber) URLStr() string {
	return s.srv.URL
}

// AcceptN consumes and returns up to n requests that were sent to the server
// before maxWait is up. returns all of the requests received. each individual
// request returned will not be returned by this func again.
func (s *Subscriber) AcceptN(n int, maxWait time.Duration) []*ReceivedRequest {
	var ret []*ReceivedRequest
	tmr := time.After(maxWait)
	for i := 0; i < n; i++ {
		select {
		case recvReq := <-s.ch:
			ret = append(ret, recvReq)
		case <-tmr:
			return ret
		}
	}
	return ret
}
