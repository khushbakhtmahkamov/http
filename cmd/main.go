package main

import (

	"github.com/khushbakhtmahkamov/http/pkg/server"

	"log"
	"net"
	"os"
	"strconv"
)
func main(){
	host := "0.0.0.0"
	port := "8080"

	if err := execute(host, port); err != nil{
		os.Exit(1);
	}
}

func execute(host string, port string)error{
	srv := server.NewServer(net.JoinHostPort(host, port))
	body := "hello";
	srv.Register("/api/category{category}/{id}", func(req *server.Request){
		_, err := req.Conn.Write([]byte(
			"HTTP/1.1 200 OK\r\n" +
				"Content-Length: " + strconv.Itoa(len(body)) + "\r\n" +
				"Content-Type: text/html\r\n" +
				"Connection: close\r\n" +
				"\r\n" + body,
		))
		if err != nil {
			log.Print(err)
		}
	})
	// body, err := ioutil.ReadFile("static/index.html");
	// if(err != nil){
	// 	log.Print("can't read file");
	// 	return err;
	// }
	// srv.Add("/", string(body))
	// srv.Add("/about", "About Golang Academy")
	return srv.Start()
}