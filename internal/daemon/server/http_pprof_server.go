package server

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)


type ServerConfig struct {
	addr string
	log *slog.Logger
}


func StartPprofServer(s ServerConfig) {
	go func() {
		if err := http.ListenAndServe(s.addr,nil); err != nil{
			s.log.Error("Failed to start pprof server")
			return
		}
	}()
}