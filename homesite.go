package main

import (
    "net/http"
    "log"
    "html/template"
    "os"

    "github.com/julienschmidt/httprouter"
)

type Block struct {
    Title string
    Body []byte
}

func (b *Block) save() error {
    return os.WriteFile(b.path(), b.Body, 0600)
}

func (b *Block) path() string {
    return "Blocks/" + b.Title + ".txt"
}

func loadBlock(title string) (*Block, error) {
    filename := "Blocks/" + title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Block{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Block) {
    t, _ := template.ParseFiles("Templates/" + tmpl + ".html")
    t.Execute(w, p)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    renderTemplate(w, "index", nil)
}

func Edit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("title")
    b, err := loadBlock(title)
    if err != nil {
        log.Println(title + ".txt not found")
        b = &Block{Title: "log"}
    }
    renderTemplate(w, "edit", b)
}

func Save(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    title := ps.ByName("title")
    b, err := loadBlock(title)
    if err != nil {
        return
    }
    b.Body = []byte(r.PostFormValue("body"))
    b.save()
    http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
    router := httprouter.New()
    router.GET("/", Index)
    router.GET("/edit/:title", Edit)
    router.POST("/save/:title", Save)

    log.Fatal(http.ListenAndServe(":8080", router))
}