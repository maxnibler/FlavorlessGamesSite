package main

import (
    "net/http"
    "log"
    "html/template"
    "os"
    "encoding/json"
    "encoding/gob"
    // "strconv"
    // "bytes"
    // "fmt"

    "github.com/gorilla/sessions"
    "github.com/julienschmidt/httprouter"
)

var tmpl = map[string]*template.Template{}
var store *sessions.CookieStore
var accounts []*Account
var users []*User


type Message struct {
    Text string
    Warning bool
    Error bool
    Success bool
}

// Utils

func loadJSON(path string) ([]map[string]string, error) {
    file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
        return nil, err
    }
    defer file.Close()
    var data []map[string]string
    if err := json.NewDecoder(file).Decode(&data); err != nil {
        log.Fatal(err)
        return nil, err
    }
    return data, nil
}

// Blocks

func blockPath(name string) string {
    return "Templates/Blocks/" + name + ".html"
}

func loadBlock(name string) *template.Template {
    path := blockPath(name)
    if tmpl[name] == nil {
        tmpl[name] = template.Must(template.ParseFiles(path))
    }
    tmpl[name] = template.Must(template.ParseFiles(path))
    return tmpl[name]
}

func Header(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    user, err := sessionUser(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    t := loadBlock("header")
    t.Execute(w, user)
}

func sendMessage(w http.ResponseWriter, message *Message, status int) {
    t := loadBlock("message")
    w.WriteHeader(status)
    t.Execute(w, message)
}

// Pages

func pagePath(name string) string {
    return "Templates/Pages/" + name + ".html" 
}

func loadPage(name string) *template.Template {
    pp := pagePath(name)
    bp := pagePath("base")
    if tmpl[name] == nil {
        // log.Printf("Initializing Template: %s", name)
        tmpl[name] = template.Must(template.ParseFiles(pp, bp))
    }
    tmpl[name] = template.Must(template.ParseFiles(pp, bp))
    return tmpl[name]
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t := loadPage("index")
    t.ExecuteTemplate(w, "base", nil)
}

func About(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t := loadPage("about")
    t.ExecuteTemplate(w, "base", nil)
}

func loadEnvs() {
    // Load Session Key
    sessionKey, err := os.ReadFile("Data/.secret/session_key.txt")
    if err != nil {
        log.Printf("Session Key not found, sessions will not work")
    } else {
        os.Setenv("SESSION_KEY", string(sessionKey))
        store = sessions.NewCookieStore(sessionKey)
    }
}

func main() {
    router := httprouter.New()

    loadEnvs()
    loadAccounts()
    loadUsers()

    gob.Register(User{})

    router.GET("/", Index)
    router.GET("/about", About)
    router.GET("/header", Header)
    router.GET("/users/login", LoginPage)
    router.POST("/users/login", LoginPost)
    router.GET("/users/login/form", LoginForm)
    router.POST("/users/signup", SignupPost)
    router.GET("/users/signup/form", SignupForm)
    router.GET("/users/profile", profilePage)
    router.GET("/blurb/:key/edit", BlurbEdit)
    router.POST("/blurb/:key/save", BlurbSave)
    router.GET("/blurb/:key", BlurbGet)
    router.GET("/team", TheTeam)
    router.GET("/team/:key", MemberBlock)
    // router.GET("/message/:key", MessageGet)

    router.ServeFiles("/static/*filepath", http.Dir("static"))

    log.Fatal(http.ListenAndServe(":8080", router))
}