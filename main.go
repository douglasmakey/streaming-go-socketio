package main

import (
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"strconv"
	"regexp"
)

//Declaramos un tipo transmitter que tendra la estructura del emisor.
type transmitter struct {
	Id		string				//Id del socket
	so		socketio.Socket			//Socket
}

//Declaramos un tipo consumer que tendra la estructura del consumidor
type consumer struct {
	Id		string		//Id del socket
	name		string		//Nommbre del consumidor
}

//Creamos el tipo namespace que tendra la estructura del mismo.
type namespace struct {
	name			string				//Nombre del namespace
	counter			int				//Consumidores conectados
	emitter 		*transmitter			//El emisor de dicho Namespace
	consumers 		map[string]*consumer		//Map para que recibe un puntero de la estructura consumidor, para almacenar los consumidores
}
//Creamos un map que recibe un puntero de la estrucutra namespacee para guardar los mismos
var namespaces = make(map[string]*namespace)

//Declaramos la url base del proyecto
var urlBase string = "http://localhost:5000/consume.html?"

func main() {
	//Iniciamos el socket
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		log.Println("On connection")

		//Declaramos la variable donde almacenaremos un puntero al Namespace
		var nsp *namespace

		//Capturamos la variable 'type' que envian al conectarse al socket por QueryParam
		tp := so.Request().FormValue("type")
		//Capturamos al nombre del namespace del emisor
		name := so.Request().FormValue("namespace")
		//Seteamos por defecto si no vienen variables.
		if tp == "" {
			tp = "consumer"
		}
		if name == ""  {
			name = "default"
		}

		//Verificamos el tipo
		if tp == "consumer" {
			//Obtenemos el namespace de acuerdo al nombre
			nsp = namespaces[name]
			//Validamos que existe el Namespace, si no desconectamos el socket actual
			if nsp == nil {
				log.Println("Namespace no encontrado")
				so.Emit("disconnect", "Desconectado")
				return
			}

			log.Println("Se ha connectado un nuevo Consumidor al Namespace: " + nsp.name)
			//Capturamos al nombre del usuario consumidor
			user := so.Request().FormValue("user")
			//Si el consumidor no envio su nombre, le asignamos uno.
			if user == "" {
				user = "Consumidor" + strconv.Itoa(nsp.counter)
			}

			//Agregamos el consumidor al MAP de consumidores en el Namespace.
			nsp.consumers[so.Id()] = &consumer{so.Id(), user}

			//Ingresamos a la sala correspondiente al namespace
			so.Join("stream-" + nsp.name)

			//Llevamos el conteo de cuantos consumidores conectados y notificamos al que emite.
			nsp.counter += 1

			//Validamos que emit no este vacio y notificamos al emisor la cantidad de consumidores.
			if nsp.emitter != nil {
				nsp.emitter.so.Emit("count-consume", nsp.counter)
			}
		} else {
			//Type Emisor

			//Creamos el namespace
			nsp = &namespace{name, 0, &transmitter{so.Id(), so}, make(map[string]*consumer)}

			//Guardamos el Namespace en el map de Namespaces
			namespaces[nsp.name] = nsp

			//Log
			log.Println("Se ha connectado un emisor y se crea el namespace: " + name)

			//Ingresamos a la sala correspondiente al namespace
			so.Join("stream-" + name)

			//Definimos la urlBase para los consumidores
			url := urlBase + "namespace=" + name

			//Emitimos al Emisor su url para consumir
			so.Emit("url", url)
		}

		//Recibimos la emicion y la enviamos a todos los consumidores correspondientes al namespace
		so.On("stream", func(image string) {
			eventAndBro := "stream-" + nsp.name
			so.BroadcastTo(eventAndBro, eventAndBro, image)
		})

		//Recibimos los mensajes del chat y reenviamos a la sala a cual pertenece
		so.On("chat", func(m string) {
			//Validamos que el mensaje no contenga etiquetas HTML
			if m, _ := regexp.MatchString(`<(\w+)((?:\s+\w+(?:\s*=\s*(?:(?:"[^"]*")|(?:'[^']*')|[^>\s]+))?)*)\s*(\/?)>`, m); !m {
				///<(\w+)((?:\s+\w+(?:\s*=\s*(?:(?:'[^']*')|(?:'[^']*')|[^>\s]+))?)*)\s*(\/?)>/
				return false
			}

			//userName para guardar el nombre de quien emite.
			var userName string
			//Tipo: 'Emisor' se guarda el nombre del Namespace
			userName = nsp.name

			//Tipo: 'consumer' se guarda el nombre del consumidor
			if tp == "consumer"{
				userName = nsp.consumers[so.Id()].name
			}

			//Creamos la data que enviaremos.
			data := make(map[string]interface{})
			data["name"] = userName
			data["message"] = m

			//Emit
			so.BroadcastTo("stream-" + nsp.name, "message-" + nsp.name, data)

		})

		//Manejamos la desconeciones
		so.On("disconnection", func() {
			if tp == "consumer"{
				log.Println("Se ha desconectado un Consumidor")
				//disminuimos el contador
				nsp.counter -= 1

				//Validamos que emisor no este vacio y notificamos al mismo la cantidad de consumidores.
				if nsp.emitter != nil {
					nsp.emitter.so.Emit("count-consume", nsp.counter)
				}
			} else {
				//Debemos eliminar el emisor y notificar a los consumidores del mismo
				log.Println("Se ha desconectado el emisor del namespace: " + nsp.name)
				so.BroadcastTo("stream-" + nsp.name, "streaming-closed", "closed")
			}

		})

	})

	//Imprimimos los errores del socket en caso que hayan.
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error: ", err)
	})

	http.Handle("/socket.io/", server)

	//Urilizamos http.FileServer y le pasamos la carpeta donde estan los archivos Estaticos.
	http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Serving at localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
