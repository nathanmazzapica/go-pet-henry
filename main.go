package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
)


var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)


	data := struct {
		Title string
		User string
	}{
		Title: "Welcome Page",
		User: "Nathan",
	}

	tmpl := template.Must(template.ParseFiles("templates/template.html"))

	tmpl.Execute(w, data)

}

func main() {
	fmt.Println("Hello, Go!")

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/",serveHome)

	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		fmt.Println("something went horribly wrong", err)
	}

}
