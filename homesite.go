package main

import (
    "net/http"
    "log"
    "html/template"
    "os"

    "github.com/julienschmidt/httprouter"
)

type Blurb struct {
    Title string
    Body []byte
}

func (b *Blurb) save() error {
    return os.WriteFile(b.path(), b.Body, 0600)
}

func (b *Blurb) path() string {
    return "Blurbs/" + b.Title + ".txt"
}

func loadBlurb(title string) (*Blurb, error) {
    filename := "Blurbs/" + title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Blurb{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Blurb) {
    t, _ := template.ParseFiles("Templates/" + tmpl + ".html")
    t.Execute(w, p)
}

func renderPage(w http.ResponseWriter, tmpl string) {
    t, _ := template.ParseFiles("Pages/" + tmpl + ".html")
    t.Execute(w, nil)
}

func templatePath(name string) string {
    return "Templates/" + name + ".html"
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    renderTemplate(w, "index", nil)
}

func Edit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("title")
    b, err := loadBlurb(title)
    if err != nil {
        b = &Blurb{Title: "log"}
    }
    renderTemplate(w, "edit", b)
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("title")
    b, err := loadBlurb(title)
    if err != nil {
        return
    }
    b.Body = []byte(r.PostFormValue("body"))
    b.save()
    tmpl, _ := template.ParseFiles(templatePath("blurb"))
    tmpl.Execute(w, b)
}

func GetBlurb(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("title")
    b, err := loadBlurb(title)
    if err != nil {
        b = &Blurb{Title: "Blurb not found"}
    }
    tmpl, _ := template.ParseFiles(templatePath("blurb"))
    tmpl.Execute(w, b)
}

func Page(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("page_name")
    renderPage(w, title)
}

func Header()

func main() {
    router := httprouter.New()
    router.GET("/", Index)
    router.GET("/edit/:title", Edit)
    router.POST("/save/:title", Save)
    router.GET("/blurb/:title", GetBlurb)
    router.GET("/page/:page_name", Page)

    log.Fatal(http.ListenAndServe(":8080", router))
}