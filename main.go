package main

import (
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
)


var consumeConnectCounter uint8

func main() {
	//Iniciamos el socket
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(so socketio.Socket) {
		log.Println("On connection")

		//Ingresamos a la sala 'stream'
		so.Join("stream")

		//Llevamos el conteo de cuantos consumidores conectados y notificamos al que emite.
		consumeConnectCounter += 1

		//Emitimos
		if consumeConnectCounter > 1 {
			log.Println("Se ha connectado un nuevo Consumidor")
			so.Emit("count-consume", consumeConnectCounter)
		}

		//Recibimos la emicion y la enviamos a todos los consumidores
		so.On("stream", func(image string) {
			so.BroadcastTo("stream", "stream", image)
		})

		so.On("disconnection", func() {
			log.Println("Consumidor disconnect")
			consumeConnectCounter -= 1
		})
	})

	//Imprimimos los errores del socket en caso que hayan.
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error: ", err)
	})

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Serving at localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
