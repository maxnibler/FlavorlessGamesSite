package main

import (
    "net/http"
    "log"
    "html/template"
    "os"
    // "fmt"

    "github.com/gorilla/sessions"
    "github.com/julienschmidt/httprouter"
)

var tmpl = map[string]*template.Template{}
var store *sessions.CookieStore

type User struct {
    Name string
    Email string
    Admin bool
}

type Blurb struct {
    Title string
    Body []byte
}

func (b *Blurb) save() error {
    path := "Data/Blurbs/" + b.Title + ".txt"
    return os.WriteFile(path, b.Body, 0600)
}

// Users

func sessionUser(w http.ResponseWriter, r *http.Request) (*User, error) {
    session, err := store.Get(r, "user-session")
    if err != nil {
		return nil, err
    }
    err = session.Save(r, w)
	if err != nil {
		return nil, err
	}
    user, yes := session.Values["user"].(*User)
    if yes {
        return user, nil
    }
    return nil, nil
}

func LoginPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    log.Printf("%s, %s", r.PostFormValue("email"), r.PostFormValue("password"))
    w.Header().Set("HX-Redirect", "/")
    w.WriteHeader(http.StatusOK)
}

// Blurbs

func loadBlurb(title string) (*Blurb, error) {
    filename := "Data/Blurbs/" + title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Blurb{Title: title, Body: body}, nil
}

func BlurbSave(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("key")
    b, err := loadBlurb(title)
    if err != nil {
        return
    }
    b.Body = []byte(r.PostFormValue("body"))
    b.save()
    t := loadBlock("blurb")
    t.Execute(w, b)
}

func BlurbGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    b, err := loadBlurb(key)
    if err != nil {
        b = &Blurb{Title: key}
    }
    t := loadBlock("blurb")
    t.Execute(w, b)
}

func BlurbEdit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("key")
    b, err := loadBlurb(title)
    if err != nil {
        b = &Blurb{Title: title}
    }
    t := loadBlock("blurbEdit")
    t.Execute(w, b)
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
    t := loadBlock("header")
    t.Execute(w, nil)
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
    _, err := sessionUser(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    t := loadPage("index")
    t.ExecuteTemplate(w, "base", nil)
}

func About(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t := loadPage("about")
    t.ExecuteTemplate(w, "base", nil)
}

func LoginPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t := loadPage("login")
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

    router.GET("/", Index)
    router.GET("/about", About)
    router.GET("/header", Header)
    router.GET("/users/login", LoginPage)
    router.POST("/users/login", LoginPost)
    router.GET("/blurb/:key/edit", BlurbEdit)
    router.POST("/blurb/:key/save", BlurbSave)
    router.GET("/blurb/:key", BlurbGet)

    router.ServeFiles("/static/*filepath", http.Dir("static"))

    log.Fatal(http.ListenAndServe(":8080", router))
}