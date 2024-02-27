package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, mux *http.ServeMux, ctx context.Context) error {
	//portStr := strconv.Itoa(port)
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		<-ctx.Done()

		ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			_ = fmt.Errorf("Run: s.httpServer.Shutdown(ctx):  %v", err)
			return
		}
	}()

	return s.httpServer.ListenAndServe()
}
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
