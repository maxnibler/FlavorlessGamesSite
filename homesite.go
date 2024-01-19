package main

import (
    "net/http"
    "log"
    "html/template"
    "os"

    "github.com/julienschmidt/httprouter"
)

var tmpl = map[string]*template.Template{}

// Blurbs

type Blurb struct {
    Title string
    Body []byte
}

func (b *Blurb) save() error {
    path := "Data/Blurbs/" + b.Title + ".txt"
    return os.WriteFile(path, b.Body, 0600)
}

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

func main() {
    router := httprouter.New()

    router.GET("/", Index)
    router.GET("/about", About)
    router.GET("/header", Header)
    router.GET("/blurb/:key/edit", BlurbEdit)
    router.POST("/blurb/:key/save", BlurbSave)
    router.GET("/blurb/:key", BlurbGet)

    log.Fatal(http.ListenAndServe(":8080", router))
}