package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    
    "github.com/gorilla/mux"

    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
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


func main() {
    initGlobals()
    globalRoom = InitCoordinator()
    RunCoordinator()
    initRouter()
    addRoutes()
    startServer()
}
