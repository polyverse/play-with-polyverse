package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/gorilla/mux"
	"github.com/polyverse/play-with-polyverse/config"
	"github.com/polyverse/play-with-polyverse/router"
	"github.com/shirou/gopsutil/load"
	"github.com/urfave/negroni"
)

func director(protocol router.Protocol, host string) (*router.DirectorInfo, error) {
	info, err := router.DecodeHost(host)
	if err != nil {
		return nil, err
	}

	port := info.Port

	if info.EncodedPort > 0 {
		port = info.EncodedPort
	}

	i := router.DirectorInfo{}
	if port == 0 {
		if protocol == router.ProtocolHTTP {
			port = 80
		} else if protocol == router.ProtocolHTTPS {
			port = 443
		} else if protocol == router.ProtocolSSH {
			port = 22
			i.SSHUser = "root"
			i.SSHAuthMethods = []ssh.AuthMethod{ssh.Password("root")}
		} else if protocol == router.ProtocolDNS {
			port = 53
		}
	}

	t, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", info.InstanceIP, port))
	if err != nil {
		return nil, err
	}
	i.Dst = t
	return &i, nil
}

func main() {
	config.ParseFlags()

	ro := mux.NewRouter()
	ro.HandleFunc("/ping", ping).Methods("GET")
	n := negroni.Classic()
	n.UseHandler(ro)

	httpServer := http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           n,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go httpServer.ListenAndServe()

	r := router.NewRouter(director)
	r.ListenAndWait(":443", ":53", ":22")
	defer r.Close()
}

func ping(rw http.ResponseWriter, req *http.Request) {
	// Get system load average of the last 5 minutes and compare it against a threashold.

	a, err := load.Avg()
	if err != nil {
		log.Println("Cannot get system load average!", err)
	} else {
		if a.Load5 > config.MaxLoadAvg {
			log.Printf("System load average is too high [%f]\n", a.Load5)
			rw.WriteHeader(http.StatusInsufficientStorage)
		}
	}

	fmt.Fprintf(rw, `{"ip": "%s"}`, config.L2RouterIP)
}
