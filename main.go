package main

import (
	"core"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/jasonlvhit/gocron"
)

var addresses map[string]int64

func init() {
	addresses = make(map[string]int64, 0)
}

func addIPv6Address(address string) {
	if _, ok := addresses[address]; !ok {
		var log = fmt.Sprintf("[%s] address added to cache map", address)
		fmt.Println(log)
	}
	addresses[address] = time.Now().Unix()
}

func deleteIPv6Address(address string) {
	var log = fmt.Sprintf("[%s] address deleted from cache map", address)
	fmt.Println(log)
	delete(addresses, address)
}

func deleteIPv6AddressIfNeeded() {
	for k, addTime := range addresses {
		var currentTime = time.Now().Unix()
		var elapsedTime = currentTime - addTime
		if elapsedTime >= 120 {
			deleteIPv6Address(k)
			if err := deleteIPv6AddrToInterface(k); err != nil {
				fmt.Println(fmt.Sprintf("An error returned while removing %s address. Error: %s", k, err.Error()))
			}
		}
	}
}

func scheduleJob() bool {
	gocron.Every(5).Seconds().Do(deleteIPv6AddressIfNeeded)
	return <-gocron.Start()
}

func main() {
	var err error
	if err = core.ReadConfig(); err != nil {
		fmt.Println("Could not read config file.")
		return
	}
	fmt.Println("Config file loaded.")

	var config = core.GetConfig()

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				http.Error(w, "HTTP requests doesn't supported yet.", http.StatusBadRequest)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	gocron.Every(5).Seconds().Do(deleteIPv6AddressIfNeeded)

	select {
	case stopped := <-gocron.Start():
		if stopped {
			fmt.Println("Scheduler stopped")
		}
	default:
		break
	}

	fmt.Println(fmt.Sprintf("%s:%d proxy started.", config.Proxy.Host, config.Proxy.Port))
	log.Fatal(server.ListenAndServe())
}

func addIPv6AddrToInterface(localAddr string) error {
	var config = core.GetConfig()
	var c = exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%d", localAddr,config.Proxy.SubnetMask), "dev", config.Proxy.Interface)
	var err error
	if err = c.Run(); err != nil {
		return err
	}
	return nil
}

func deleteIPv6AddrToInterface(localAddr string) error {
	var config = core.GetConfig()
	var c = exec.Command("ip", "addr", "del", fmt.Sprintf("%s/%d", localAddr,config.Proxy.SubnetMask), "dev", config.Proxy.Interface)
	var err error
	if err = c.Start(); err != nil {
		return err
	}
	return nil
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	localAddr := r.Header.Get("x-proxy-ip")

	if localAddr == "" {
		var config = core.GetConfig()
		localAddr = config.Proxy.TempIP
	}

	addIPv6Address(localAddr)
	addIPv6AddrToInterface(localAddr)

	localTCPAddr, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf("[%s]:0", localAddr))
	if err != nil {
		http.Error(w, "Given proxy IP does not seems to be resolved", http.StatusBadRequest)
		return
	}

	var dialReady = false
	var d *net.Dialer
	var dst net.Conn

	for dialReady == false {
		d = &net.Dialer{LocalAddr: localTCPAddr, Timeout: 30 * time.Second}
		if dst, err = d.Dial("tcp6", r.Host); err == nil {
			dialReady = true
		}
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(dst, conn)
	go transfer(conn, dst)
}

func transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	_, _ = io.Copy(dst, src)
}
