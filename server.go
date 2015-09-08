package testsrv

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

// Server is a HTTP server that you can send real HTTP requests to in your unit tests.
// Each request is handled by a http.Handler that you specify, and the server records each
// request after your handler finishes execution. You can later retreive details about each
// received request
type Server struct {
	closed   int64
	closeSig chan struct{} // closed when Close is called
	ch       chan *ReceivedRequest
	srv      *httptest.Server
}

// StartServer starts the server listening in the background. All requests sent to the server
// at s.URLStr() are sent to handler, flushed and then recorded for later retrieval using the AcceptN func
func StartServer(handler http.Handler) *Server {
	ch := make(chan *ReceivedRequest)
	closeSig := make(chan struct{})
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := uuid.New()
		w.Header().Add(UUIDHeaderName, u)
		handler.ServeHTTP(w, r)
		req := &ReceivedRequest{Request: r, Time: time.Now(), UUID: u}
		select {
		case <-closeSig:
		case ch <- req:
		}
	})

	s := httptest.NewServer(wrapper)
	return &Server{
		closed:   0,
		closeSig: closeSig,
		ch:       ch,
		srv:      s,
	}
}

// Close releases all resources on this subscriber
func (s *Server) Close() {
	if atomic.CompareAndSwapInt64(&s.closed, 0, 1) {
		s.srv.Close()
		s.srv.CloseClientConnections()
		close(s.closeSig)
		return
	}
}

// URLStr returns the string representation of the URL to call to hit this server
func (s *Server) URLStr() string {
	return s.srv.URL
}

// AcceptN consumes and returns up to n requests that were sent to the server
// before maxWait is up. returns all of the requests received. each individual
// request returned will not be returned by this func again.
func (s *Server) AcceptN(n int, maxWait time.Duration) []*ReceivedRequest {
	var ret []*ReceivedRequest
	tmr := time.After(maxWait)
	for i := 0; i < n; i++ {
		select {
		case recvReq := <-s.ch:
			ret = append(ret, recvReq)
		case <-s.closeSig:
			return ret
		case <-tmr:
			return ret
		}
	}
}
