package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func g6Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "www/g6_launcher/index.html")

}

// serveLauncherHTML is responsible for serving the launcher HTML and (HACK) serverlist.xml.
func serveLauncherHTML(listenAddr string) {
	// Manually route the folder root to index.html? Is there a better way to do this?
	router := httprouter.New()
	router.GET("/g6_launcher/", g6Index)

	static := httprouter.New()
	static.ServeFiles("/*filepath", http.Dir("www"))
	router.NotFound = static

	http.ListenAndServe(listenAddr, router)
}
