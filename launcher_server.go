package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
)

func g6Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "www/g6_launcher/index.html")

}

func serverUniqueName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(dump))

	fmt.Fprintf(w, `<?xml version="1.0" encoding="ISO-8859-1"?><uniq code="200">OK</uniq>`)
}

// serveLauncherHTML is responsible for serving the launcher HTML and (HACK) serverlist.xml.
func serveLauncherHTML(listenAddr string) {
	// Manually route the folder root to index.html? Is there a better way to do this?
	router := httprouter.New()
	router.GET("/g6_launcher/", g6Index)
	router.GET("/server/unique.php", serverUniqueName)

	static := httprouter.New()
	static.ServeFiles("/*filepath", http.Dir("www"))
	router.NotFound = static

	http.ListenAndServe(listenAddr, handlers.LoggingHandler(os.Stdout, router))
}
