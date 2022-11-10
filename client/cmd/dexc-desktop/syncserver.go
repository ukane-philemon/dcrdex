// This code is available on the terms of the project LICENSE.md file,
// also available online at https://blueoakcouncil.org/license/1.0.0.

/*
The sync server is a tiny HTTP server that handles synchronization of multiple
instances of dexc-desktop. Any dexc-desktop instance that manages to get a
running webserver will create a sync server and publish the address at specific
file location. On startup, before getting too far, dexc-desktop looks for a
file, reads the address, and attempts to make contact. If contact is made,
the requesting dexc-desktop will exit without error. The receiving dexc-desktop
will reopen the window if it is closed.

The sync server also enables killing a dexc-desktop process that is running
in the background because of active orders.
*/

package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	syncFilename = "syncfile"
)

// getSyncAddress: Get the address stored in the syncfile. If the file doesn't
// exist, an empty string is returned with no error.
func getSyncAddress(syncDir string) (addr string, err error) {
	syncFile := filepath.Join(syncDir, syncFilename)
	addrB, err := os.ReadFile(syncFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		return "", nil
	}
	return string(addrB), nil
}

// sendKillSignal yeets a kill command to any address in the sync file.
func sendKillSignal(syncDir string) {
	addr, err := getSyncAddress(syncDir)
	if err != nil {
		log.Errorf("Error getting sync address for kill signal: %v", err)
		return
	}
	if addr == "" {
		log.Errorf("No sync address found for kill signal: %v", err)
		return
	}
	resp, err := http.Get("http://" + addr + "/kill")
	if err != nil {
		log.Errorf("Error sending kill signal to sync server: %v", err)
		return
	}
	if resp == nil {
		log.Errorf("nil kill response")
		return
	}
	if resp.StatusCode == http.StatusOK {
		log.Errorf("Unexpected response code from kill signal send: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// synchronize attempts to determine the state of any other dexc-desktop
// instances that are running. If none are found, startServer will be true.
// If another instance is found, close will be true, and this instance of
// dexc-desktop should exit immediately.
func synchronize(syncDir string) (startServer, close bool, err error) {
	addr, err := getSyncAddress(syncDir)
	if err != nil {
		return false, false, err
	}
	if addr == "" {
		// Just start the server then.
		return true, false, nil
	}
	resp, err := http.Get("http://" + addr)
	if err == nil && resp.StatusCode == http.StatusOK {
		// Other instance will open the window.
		return false, true, nil
	}
	return true, false, nil
}

// runServer runs an instance of the sync server. Received commands are
// communicated via unbuffered channels. Blocking channels are ignored.
func runServer(ctx context.Context, syncDir string, openC chan<- struct{}, killC chan<- os.Signal) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Errorf("ResolveTCPAddr error: %v", err)
		return
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Errorf("ListenTCP error: %v", err)
		return
	}

	f, err := os.Create(filepath.Join(syncDir, syncFilename))
	if err != nil {
		log.Errorf("Failed to start the sync server: %v", err)
		return
	}
	// Write the address to the syncfile.
	if _, err := f.Write([]byte(l.Addr().String())); err != nil {
		log.Errorf("Error writing syncfile: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Errorf("Error closing syncfile: %v", err)
	}

	srv := &http.Server{
		Addr:    l.Addr().String(),
		Handler: &syncServer{openC, killC},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("listen: %s\n", err)
		}
		log.Infof("Sync server off")
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("http.Server Shutdown errored: %v", err)
	}
	wg.Wait()
}

type syncServer struct {
	openC chan<- struct{}
	killC chan<- os.Signal
}

func (s *syncServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	switch r.URL.Path {
	case "/":
		select {
		case s.openC <- struct{}{}:
			log.Info("Window reopened")
		default:
			log.Infof("Ignored a window reopen request from another instance")
		}
	case "/kill":
		select {
		case s.killC <- os.Interrupt:
			log.Info("Kill signal received")
		default:
			log.Infof("Ignored a window reopen request from another instance")
		}
	default:
		log.Errorf("syncServer received a request with an unknown path %q", r.URL.Path)
	}

}
