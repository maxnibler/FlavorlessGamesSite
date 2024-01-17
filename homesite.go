package main

import (
    "fmt"
    "net/http"
    "log"
    "html/template"
    "os"

    "github.com/julienschmidt/httprouter"
)

type Page struct {
    Title string
    Body []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t, _ := template.ParseFiles("Templates/index.html")
    t.Execute(w, nil)
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func Edit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    p, err := loadPage("log")
    if err != nil {
        log.Println("log.txt not found")
        p = &Page{Title: "log"}
    }
    t, _ := template.ParseFiles("Templates/edit.html")
    t.Execute(w, p)
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    log.Println(r.PostFormValue("body"))
    p, err := loadPage("log")
    if err != nil {
        return
    }
    p.Body = []byte(r.PostFormValue("body"))
    p.save()
    http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
    router := httprouter.New()
    router.GET("/", Index)
    router.GET("/hello/:name", Hello)
    router.GET("/edit/", Edit)
    router.POST("/save/", Save)

    log.Fatal(http.ListenAndServe(":8080", router))
}