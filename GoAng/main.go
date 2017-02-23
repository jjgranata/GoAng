package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"github.com/googollee/go-socket.io"
)

const chat = "/chat"

func main() {
	messageStr := []string{}
	var mut sync.Mutex
	runtime.GOMAXPROCS((runtime.NumCPU() * 2) + 1)

	socketObj, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	socketObj.On("connection", func(sockioInst socketio.Socket) {
		var handle string
		handle = "User-" + sockioInst.Id()
		log.Println("on connection", handle)
		sockioInst.Join(chat)

		mut.Lock()
		for i, _ := range messageStr {
			sockioInst.Emit("message", messageStr[i])
		}
		mut.Unlock()

		sockioInst.On("joined_message", func(m string) {
			handle = m
			log.Println("joined log ", m)
			resolutionMessage := map[string]interface{}{
				"username": handle,
				"dateTime": time.Now().UTC().Format(time.RFC3339Nano),
				"type":     "joined_message",
			}
			jsonRes, _ := json.Marshal(resolutionMessage)
			sockioInst.Emit("message", string(jsonRes))
			sockioInst.BroadcastTo(chat, "message", string(jsonRes))
		})
		sockioInst.On("send_message", func(m string) {
			log.Println("send log ", handle)
			resolutionSend := map[string]interface{}{
				"username": handle,
				"message":  m,
				"dateTime": time.Now().UTC().Format(time.RFC3339),
				"type":     "message",
			}
			jsonRes, _ := json.Marshal(resolutionSend)
			mut.Lock()
			if len(messageStr) == 100 {
				messageStr = messageStr[1:100]
			}
			messageStr = append(messageStr, string(jsonRes))
			mut.Unlock()
			sockioInst.Emit("message", string(jsonRes))
			sockioInst.BroadcastTo(chat, "message", string(jsonRes))
		})
		sockioInst.On("disconnection", func() {
			log.Println("disconnected ", handle)
		})
	})
	socketObj.On("error", func(so socketio.Socket, err error) {
		log.Println(err)
	})

	http.Handle("/socket.io/", socketObj)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", http.FileServer(http.Dir("./templates/")))

	var l string = os.Getenv("LISTEN")

	if l == "" {
		l = ":5000"
	}

	log.Println("port ", l)
	http.ListenAndServe(l, nil)
}