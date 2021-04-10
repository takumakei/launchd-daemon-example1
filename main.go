package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/takumakei/launchd-daemon-example1/launch"
)

// SocketName is defined in plist.
const SocketName = "MySocketName"

func main() {
	initLog()
	log.Printf("[INFO] com.takumakei.example1 start pid=%d\n", os.Getpid())
	defer log.Printf("[INFO] com.takumakei.example1 exit pid=%d\n", os.Getpid())

	debugPrintEnvs()

	// Retrieve listeners managed by launchd.
	ll, err := launch.SocketListeners(SocketName)
	if err != nil {
		log.Printf("[ERROR] launch_socket_server: %v\n", err)
		os.Exit(1)
	}

	// Attach http service to each listeners.
	ss := startServers(ll)

	// Wait timeout or a signal.
	//
	// Should not exit in less than 10 seconds
	//
	// https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
	//
	// > Important: If your daemon shuts down too quickly after being launched,
	// > launchd may think it has crashed. Daemons that continue this behavior may
	// > be suspended and not launched again when future requests arrive. To avoid
	// > this behavior, do not shut down for at least 10 seconds after launch.
	waitSignalOrTimeout(16 * time.Second)

	// Shutdown http services.
	shutdown(ss)
}

func initLog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.SetPrefix(fmt.Sprintf("[%d]", os.Getpid()))
	log.SetOutput(os.Stdout)
}

func debugPrintEnvs() {
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		for len(kv) < 2 {
			kv = append(kv, "")
		}
		log.Printf("[DEBUG] env[%s]=[%s]\n", kv[0], kv[1])
	}
}

func startServers(ll []net.Listener) []*http.Server {
	http.HandleFunc("/", hello)

	ss := make([]*http.Server, len(ll))
	for i, lis := range ll {
		srv := &http.Server{}
		ss[i] = srv
		go srv.Serve(lis)
	}
	return ss
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(
		res,
		"%d[%s] hello world\n",
		os.Getpid(),
		time.Now().Format(time.RFC3339Nano),
	)
}

func waitSignalOrTimeout(timeout time.Duration) {
	sigint := make(chan os.Signal)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigint)

	t := time.NewTimer(timeout)
	defer t.Stop()

	select {
	case <-t.C:
		log.Println("[INFO] timeout")
	case sig := <-sigint:
		log.Printf("[INFO] signal: %v\n", sig)
	}
}

func shutdown(ss []*http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	var g sync.WaitGroup
	g.Add(len(ss))
	for _, e := range ss {
		srv := e
		go func() {
			shutdownServer(ctx, srv)
			g.Done()
		}()
	}
	g.Wait()
}

func shutdownServer(ctx context.Context, srv *http.Server) {
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Printf("[WARN] Shutdown: %v\n", err)
	}
}
