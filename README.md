Un pequeño proyecto de Streaming en GO, utilizando Socketio

Buenas chicos, este pequeño proyecto lo lleve a cabo para practicar un poco con GO, fue un proyecto que vi hace un tiempo
por internet que hicieron con Nodejs y lo quise replicar en Go.

Utilizando la libreria Socketio en Go:
<a href="https://github.com/googollee/go-socket.io">https://github.com/googollee/go-socket.io</a>

Más las librerías estándar de Go:
* log
* net/http

El código es muy simple y la mayoría tiene comentarios que explican su funcionamiento.

<strong>- Archivo main.go</strong>
Declaramos una variable uint8 'unsigned 8-bit' para mantener el numero de consumidores conectados que va en rango (0-255)
var consumeConnectCounter uint8

Manejando los eventos:
stream: para recibir y transmitir
count-consume: para notificar el emisor cuantos consumidores estan conectados
disconnection: para llevar el control del contador de consumidores.

<pre class="lang:go decode:true">package main

import (
	"github.com/googollee/go-socket.io"
	"log"
	"net/http"
)


//Declaramos una variable, en la cual llevaremos el conteo de los consumidores conectados !
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
		if consumeConnectCounter &gt; 1 {
			log.Println("Se ha connectado un nuevo Consumidor")
			so.Emit("count-consume", consumeConnectCounter)
		}

		//Recibimos la emicion y la enviamos a todos los consumidores
		so.On("stream", func(image string) {
			so.BroadcastTo("stream", "stream", image)
		})

		so.On("disconnection", func() {
			log.Println("Se ha desconectado un Consumidor")
			consumeConnectCounter -= 1
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

</pre>

<strong>- Archivo index.html</strong>

Simple links para visualizar paginas "emit" y "consume"
<pre class="lang:xhtml decode:true">
<!DOCTYPE html>
<html lang="es">
  <head>
    <meta charset="UTF-8">
    <title>WebCam Live</title>
  </head>
  <body>
    <p>Streaming en Go con Socketio</p>
    <p><a href="emit.html">Emitir</a></p>
    <p><a href="consume.html">Consumir</a></p>
  </body>
</html>
</pre>

<strong>- Archivo emit.html</strong>

En el archivo "emit.html" el objeto menos común es 'navigator.getUserMedia'

Pide al usuario permiso para usar un dispositivo multimedia como una cámara o micrófono. Si el usuario concede este permiso,
el successCallback es invocado en la aplicación llamada con un objeto LocalMediaStream como argumento.

//En este proyecto, solo solicite 'video'
navigator.getUserMedia ( { video: true, audio: true }, successCallback, errorCallback );

<strong>successCallback</strong>
La función getUserMedia llamará a la función especificada en el successCallback con el objeto LocalMediaStream
que contenga la secuencia multimedia. Puedes asignar el objeto al elemento apropiado y trabajar con él.

<strong>errorCallback</strong>
La función getUserMedia llama a la función indicada en el errorCallback con un código como argumento.

<pre class="lang:xhtml decode:true ">
<html lang="es">
<head>
  <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <link rel="stylesheet" href="css/init.css">
  <script src="https://code.jquery.com/jquery-1.12.4.min.js"   integrity="sha256-ZosEbRLbNQzLpnKIkEdrPv7lOy9C27hHQ+Xp8a4MxAQ="   crossorigin="anonymous"></script>
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/1.4.8/socket.io.js"></script>
  <title>Emit Video</title>
</head>
<body>
    <h2 class="center">Video</h2>
    <video src="" id="video" autoplay="true"></video>
    <canvas id="preview" style="display:none;"></canvas>

    <div  id="box_logger">
        <div class="logger">
            <i id="box_logger-icon"></i>
            <spam id="logger"></spam>
        </div>
    </div>

    <div  class="isa_info">
        <div class="logger">
            <i class="fa fa-info-circle"></i>
            Consumidores conectados: <spam id="counter"></spam>
        </div>
    </div>

  <script type="text/javascript" charset="utf-8">
    //Obtenemos el canvas y el contexto y fijamos su tamaño.
    var canvas = document.getElementById("preview");
    var context = canvas.getContext("2d");
    canvas.width = 800;
    canvas.height = 600;
    context.width = canvas.width;
    context.height = canvas.height;

    //Obtenemos el DIV donde se mostrara el video
    var video = document.getElementById("video");

    //Instanciamos Socketio
    var socket = io();


    //Actualizamos el contador
    socket.on("count-consume", function (count) {
        console.log(count)
      $("#counter").text(count)
    });

    function logger(type, msg){
        if (type == 'sucess'){
            $("#box_logger").addClass("isa_success")
            $("#box_logger-icon").addClass("fa fa-check")
        }else{
            $("#box_logger").addClass("isa_error")
            $("#box_logger-icon").addClass("fa fa-times-circle")
        }
      $("#logger").text(msg);
    }

    function ok(stream){
      video.src = window.URL.createObjectURL(stream)
      logger('sucess', 'Camara disponible !');
    }

    function fail(){
      logger('error', 'Ha ocurrido un problema, No logramos detectar su Camara !');
    }


    function Consume(video, context){
      context.drawImage(video, 0, 0, context.width, context.height);
      socket.emit('stream', canvas.toDataURL('image/webp'));
    }

    $(function(){
        //Obtenemos el getUserMedia segun el navegador
        navigator.getUserMedia = (navigator.getUserMedia || navigator.webkitGetUserMedia
                                || navigator.mozGetUserMedia || navigator.msgGetUserMedia);

      if(navigator.getUserMedia){
        navigator.getUserMedia({video:true}, ok, fail);
      }

      setInterval(function(){
        Consume(video, context)
      }, 50);

    });
  </script>
</body>
</html>
</pre>

<strong>- Archivo consume.html</strong>

<pre class="lang:xhtml decode:true ">
<html>
<head>
  <meta http-equiv="content-type" content="text/html; charset=utf-8">
  <link rel="stylesheet" href="css/init.css">
  <script type="text/javascript" src="https://cdnjs.cloudflare.com/ajax/libs/socket.io/1.4.8/socket.io.js"></script>
  <title>Consume stream</title>
</head>
<body>
<div class="stream" style="text-align: center">
  <h2>Streaming</h2>
  <img id="play">
</div>

  <script type="text/javascript" charset="utf-8">
    var socket = io();

    socket.on('stream', function(image){
      var img = document.getElementById("play");
      img.src = image;
    });
  </script>
</body>
</html>
</pre>

<strong>- Archivo init.css</strong>
<pre lang="css">
/* Tengo como 1 año sin hacer nada de  CSS */
@import url('//maxcdn.bootstrapcdn.com/font-awesome/4.3.0/css/font-awesome.min.css');

#video {
    margin-left: auto;
    margin-right: auto;
    display: block
}

.center{
    text-align: center;
}

.logger{
    margin-left: -4px;
    margin-top: -8px;
}
.isa_info, .isa_success, .isa_error {
    height: 15px;
    width: 615px;
    padding:12px;
    margin-left: auto;
    margin-right: auto;
    display: block
}
.isa_info {
    color: #00529B;
    background-color: #BDE5F8;
}
.isa_success {
    color: #4F8A10;
    background-color: #DFF2BF;
}
.isa_error {
    color: #D8000C;
    background-color: #FFBABA;
}
.isa_info i, .isa_success i,  .isa_error i {
    font-size:2em;
    vertical-align:middle;
}
</pre>


Para los que deseen ver el repositorio y contribuir o hacer alguna crítica constructiva :D

Github: <a href="https://github.com/douglasmakey/streaming-go-socketio">https://github.com/douglasmakey/streaming-go-socketio</a>