package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allows connection from 8080.
	},
}
var clients = make(map[*websocket.Conn]bool)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrades HTTP connection to WebSocket.
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	clients[ws] = true

	// Defer runs when the function exits.
	defer func() {
		delete(clients, ws)
		ws.Close()
	}()

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}

func notifyClients() {
	for ws := range clients {
		_ = ws.WriteMessage(websocket.TextMessage, []byte("reload"))
	}
}

func waitAndNotifyClients() {
	const maxWait = 5 * time.Second
	const pollInterval = 100 * time.Millisecond

	time.Sleep(500 * time.Millisecond)

	deadline := time.Now().Add(maxWait)

	for {
		resp, err := http.Get("http://localhost:8080/")
		// Server takes some time to restart.
		// Wait until it responds or timeout.
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}

		if time.Now().After(deadline) {
			log.Println("Timeout waiting for server, reloading.")
			break
		}
		time.Sleep(pollInterval)
	}

	notifyClients()
}

func main() {
	var watchDirs string
	var addr string
	flag.StringVar(&watchDirs, "dirs", "internal/http/templates", "Comma-separated list of directories to watch.")
	flag.StringVar(&addr, "addr", ":35729", "Address to listen on.")
	flag.Parse()

	http.HandleFunc("/livereload", wsHandler)

	http.HandleFunc("/livereload.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		// Injects livereload script into the served page.
		// Any message will trigger a reload.
		w.Write([]byte(`
			(function () {
				var ws = new WebSocket("ws://" + location.hostname + ":35729/livereload");
				ws.onmessage = function () { location.reload(); };
			})();
		`))
	})

	go func() {
		log.Printf("Livereload server listening on %s", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, dir := range filepath.SplitList(watchDirs) {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				watcher.Add(path)
			}

			return nil
		})
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				log.Println("Change detected:", event.Name)
				go waitAndNotifyClients()
			}

		case err := <-watcher.Errors:
			log.Println("Watcher error:", err)
		}
	}
}
