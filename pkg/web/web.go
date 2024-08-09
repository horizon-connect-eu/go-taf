package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gorilla/websocket"
)

type socketEventType int

const (
	REGISTER = iota
	UNREGISTER
	RESPONSE
	REQUEST
	PUBLISH
)

type socketEvent struct {
	eventType socketEventType
	socket    chan socketEvent
	Data      string `json:"data"`
	Id        string `json:"id"`
}

type wsPublishWriter struct {
	queue chan socketEvent
}

func (e wsPublishWriter) Write(p []byte) (int, error) {
	e.queue <- socketEvent{eventType: PUBLISH, Data: string(p), Id: "UPDATE", socket: nil}
	return len(p), nil
}

func initHandlerThread(queue chan socketEvent) {
	sockets := []chan socketEvent{}
	var cp strings.Builder

	for evt := range queue {
		log.Println(evt)

		switch evt.eventType {
		case REGISTER:
			sockets = append(sockets, evt.socket)
			evt.socket <- socketEvent{eventType: RESPONSE, Id: "CHECKPOINT", Data: cp.String(), socket: nil}

		case UNREGISTER:
			idx := slices.IndexFunc(sockets, func(c chan socketEvent) bool { return c == evt.socket })
			if idx >= 0 {
				sockets = append(sockets[:idx], sockets[idx+1:]...)
			}

		case PUBLISH:
			cp.WriteString(evt.Data)
			for s := range sockets {
				// fan out
				sockets[s] <- socketEvent{eventType: RESPONSE, Id: "UPDATE", Data: evt.Data, socket: nil}
			}

		case REQUEST:
			// todo rpc semantics?

		case RESPONSE:
			// handler thread should not receive RESPONSEs, so we ignore them
		}
	}
}

func InitWebInterface(logger *slog.Logger, webFrontend embed.FS, webFrontendPath string) *slog.Logger {
	port := os.Getenv("WEB_INTERFACE_PORT")

	if port == "" {
		return logger
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	staticFS := fs.FS(webFrontend)
	frontendDir, err := fs.Sub(staticFS, webFrontendPath)
	if err != nil {
		panic(err)
	}

	queue := make(chan socketEvent)

	go initHandlerThread(queue)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		socket := make(chan socketEvent, 10)
		queue <- socketEvent{eventType: REGISTER, Id: "", Data: "", socket: socket}

		go func() {
			for {
				messageType, p, err := ws.ReadMessage()
				if err != nil {
					queue <- socketEvent{eventType: UNREGISTER, Id: "", Data: "", socket: socket}
					return
				}

				if messageType == websocket.BinaryMessage || messageType == websocket.TextMessage {
					var data map[string]interface{}
					if err := json.Unmarshal(p, &data); err != nil {
						continue
					}

					id, ok := data["id"].(string)
					if !ok {
						continue
					}

					queue <- socketEvent{eventType: REQUEST, Id: id, Data: string(p), socket: socket}
				}
			}
		}()

		go func() {
			for res := range socket {
				msg, err := json.Marshal(res)
				if err != nil {
					continue
				}

				if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					queue <- socketEvent{eventType: UNREGISTER, Id: "", Data: "", socket: socket}
					return
				}
			}
		}()
	})

	fileServer := http.FileServer(http.FS(frontendDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// support vue url routing by redirecting all 404 requests to root resource
		_, err := fs.Stat(frontendDir, r.URL.Path[1:])
		if err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	}))

	fmt.Fprintln(os.Stderr, "Starting web interface server: http://127.0.0.1:"+port)

	go func() {
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			panic(err)
		}
	}()

	handlerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	writer := wsPublishWriter{queue}

	weblogger := slog.New(slog.NewJSONHandler(writer, handlerOpts))

	return weblogger
}
