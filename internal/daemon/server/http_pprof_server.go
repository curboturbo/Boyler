package server

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)


type ServerConfig struct {
	Addr string
	Log *slog.Logger
}

func StartPprofServer(s ServerConfig) {
	go func() {
		if err := http.ListenAndServe(s.Addr,nil); err != nil{
			s.Log.Error("Failed to start pprof server")
			return
		}
	}()
}