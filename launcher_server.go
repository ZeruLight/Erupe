package main

import (
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	//"github.com/julienschmidt/httprouter"
)

// GetOutboundIP4 gets the preferred outbound ip4 of this machine
// From https://stackoverflow.com/a/37382208
func GetOutboundIP4() net.IP {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.To4()
}

func serverList(w http.ResponseWriter, r *http.Request) {
	// TODO(Andoryuuta): Redo launcher server to allow configurable serverlist host and port.
	fmt.Fprintf(w, `<?xml version="1.0"?><server_groups><group idx='0' nam='Erupe' ip='%s' port="53312"/></server_groups>`, GetOutboundIP4().String())

}

func serverUniqueName(w http.ResponseWriter, r *http.Request) {
	// TODO(Andoryuuta): Implement checking for unique character name.
	fmt.Fprintf(w, `<?xml version="1.0" encoding="ISO-8859-1"?><uniq code="200">OK</uniq>`)
}

func jpLogin(w http.ResponseWriter, r *http.Request) {
	// HACK: Return the given password back as the `skey` to defer the login logic to the sign server.

	resultJSON := fmt.Sprintf(`{"result": "Ok", "skey": "%s", "code": "000", "msg": ""}`, r.FormValue("pw"))

	fmt.Fprintf(w,
		`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
		<html>
		<body onload="doPost();">
		<script type="text/javascript">
		function doPost(){
			parent.postMessage(document.getElementById("result").getAttribute("value"), "http://cog-members.mhf-z.jp");
		}
		</script>
		<input id="result" value="%s"/>
		</body>
		</html>`, html.EscapeString(resultJSON))

}

func setupServerlistRoutes(r *mux.Router) {
	// TW
	twServerList := r.Host("mhf-n.capcom.com.tw").Subrouter()
	twServerList.HandleFunc("/server/unique.php", serverUniqueName) // Name checking is also done on this host.
	twServerList.HandleFunc("/server/serverlist.xml", serverList)

	// JP
	jpServerList := r.Host("srv-mhf.capcom-networks.jp").Subrouter()
	jpServerList.HandleFunc("/serverlist.xml", serverList)
}

func setupOriginalLauncherRotues(r *mux.Router) {
	// TW
	twMain := r.Host("mhfg.capcom.com.tw").Subrouter()
	twMain.PathPrefix("/").Handler(http.FileServer(http.Dir("./www/tw/")))

	// JP
	jpMain := r.Host("cog-members.mhf-z.jp").Subrouter()
	jpMain.PathPrefix("/").Handler(http.FileServer(http.Dir("./www/jp/")))

	// JP Launcher does additional auth over HTTP that the TW launcher doesn't.
	jpAuth := r.Host("www.capcom-onlinegames.jp").Subrouter()
	jpAuth.HandleFunc("/auth/launcher/login", jpLogin) //.Methods("POST")
	jpAuth.PathPrefix("/auth/").Handler(http.StripPrefix("/auth/", http.FileServer(http.Dir("./www/jp/auth/"))))

}

func setupCustomLauncherRotues(r *mux.Router) {
	// TW
	twMain := r.Host("mhfg.capcom.com.tw").Subrouter()
	twMain.PathPrefix("/g6_launcher/").Handler(http.StripPrefix("/g6_launcher/", http.FileServer(http.Dir("./www/erupe/"))))

	// JP
	jpMain := r.Host("cog-members.mhf-z.jp").Subrouter()
	jpMain.PathPrefix("/launcher/").Handler(http.StripPrefix("/launcher/", http.FileServer(http.Dir("./www/erupe"))))
}

// serveLauncherHTML is responsible for serving the launcher HTML, serverlist, unique name check, and JP auth.
func serveLauncherHTML(listenAddr string, useOriginalLauncher bool) {
	r := mux.NewRouter()

	setupServerlistRoutes(r)

	if useOriginalLauncher {
		setupOriginalLauncherRotues(r)
	} else {
		setupCustomLauncherRotues(r)
	}
	/*
		http.ListenAndServe(listenAddr, handlers.CustomLoggingHandler(os.Stdout, r, func(writer io.Writer, params handlers.LogFormatterParams) {
			dump, _ := httputil.DumpRequest(params.Request, true)
			writer.Write(dump)
		}))
	*/

	http.ListenAndServe(listenAddr, handlers.LoggingHandler(os.Stdout, r))
}
