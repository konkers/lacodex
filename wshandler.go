package lacodex

import (
	"io"
	"net/http"
	"time"

	"github.com/cskr/pubsub"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// Time allowed to write the file to the client.
var writeWait = 10 * time.Second

// Time allowed to read the next pong message from the client.
var pongWait = 60 * time.Second

// Send pings to client with this period. Must be less than pongWait.
var pingPeriod = (pongWait * 9) / 10

func wsSetTimeouts(write time.Duration, pong time.Duration) {
	writeWait = write
	pongWait = pong
	pingPeriod = (pongWait * 9) / 10
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 100,
	// TODO: Figure out propper origin checking.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsQueryFunc func(w io.Writer) error

type wsHandler struct {
	ps *pubsub.PubSub
	f  WsQueryFunc
}

func (h *wsHandler) reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			glog.Warning(err)
			break
		}
	}
}

func (h *wsHandler) writer(ws *websocket.Conn, closeC chan struct{}) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	updateC := h.ps.Sub("update")
	exitC := h.ps.Sub("exit")

	defer func() {
		go h.ps.Unsub(exitC)
		go h.ps.Unsub(updateC)
		for range exitC {
		}
		for range updateC {
		}
	}()

	ws.SetWriteDeadline(time.Now().Add(writeWait))
	w, _ := ws.NextWriter(websocket.TextMessage)
	err := h.f(w)
	if err != nil {
		return
	}
	w.Close()

L:
	for {
		select {
		case <-updateC:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := ws.NextWriter(websocket.TextMessage)
			if err != nil {
				glog.Warning(err)
				break L
			}
			err = h.f(w)
			if err != nil {
				glog.Warning(err)
				break L
			}
			w.Close()

		case <-closeC:
			break L

		case <-exitC:
			break L

		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				break L
			}
		}
	}
	glog.Info("Websocket closing.")
}

func (h *wsHandler) handler(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.URL.Query()["async"]; ok {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		closeC := make(chan struct{})
		go h.writer(ws, closeC)
		h.reader(ws)
		close(closeC)
	} else {
		w.Header().Set("Content-Type", "application/json")
		err := h.f(w)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "Can't service query: %v", err)
			return
		}
	}
}

func WsHandler(ps *pubsub.PubSub, q WsQueryFunc) http.HandlerFunc {
	h := &wsHandler{
		ps: ps,
		f:  q,
	}
	return http.HandlerFunc(h.handler)
}
