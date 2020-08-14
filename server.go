package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "os"
    "strconv"

    "github.com/gorilla/mux"

    "github.com/sdwr/multi-go/logger"
    "github.com/sdwr/multi-go/socket"
)

var router *mux.Router
var globalRoom *socket.Room
func initGlobals() {

}


//SERVER FUNCTIONS
func indexHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadFile("./web/index.html")
    fmt.Fprintf(w, "%s", body)
}
//make sure global room is init first
func socketHandler(w http.ResponseWriter, r *http.Request) {
    socket.ServeWs(globalRoom, w, r)
}

func initRouter() {
    router = mux.NewRouter()
}

func addRoutes() {
    router.HandleFunc("/socket",socketHandler)
    router.PathPrefix("/web/").Handler(http.StripPrefix("/web/",http.FileServer(http.Dir("./web/"))))
    router.HandleFunc("/", indexHandler)
}

func startServer() {
    log.Println("running server on port 4404")
    log.Fatal(http.ListenAndServe(":4404", router))
}

func setLogLevel() {
    args := os.Args[1:]
    var level int
    if len(args) > 0 {
	level, _ = strconv.Atoi(args[0])
    } else {
	level = 2
    }
    logger.InitLogger(level)
}


func main() {
    setLogLevel()
    initGlobals()
    globalRoom = InitCoordinator()
    initRouter()
    addRoutes()
    startServer()
}
