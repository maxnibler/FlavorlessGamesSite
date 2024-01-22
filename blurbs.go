package main

import (
    "net/http"
    // "log"
    // "html/template"
    "os"
    // "encoding/json"
    // "encoding/gob"
    // "strconv"
    // "bytes"
    // "fmt"
	
    "github.com/julienschmidt/httprouter"
)

type Blurb struct {
    Title string
    Body []byte
}

func (b *Blurb) save() error {
    path := "Data/Blurbs/" + b.Title + ".txt"
    return os.WriteFile(path, b.Body, 0600)
}

type BlurbUser struct {
    Blurb *Blurb
    User *User
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
    user, _ := sessionUser(w, r)
    status := http.StatusOK
    if user.Admin {
        b.save()
    } else {
        status = http.StatusUnauthorized
    }
    t := loadBlock("blurb")
    w.WriteHeader(status)
    t.Execute(w, BlurbUser{Blurb:b, User:user})
}

func BlurbGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    b, err := loadBlurb(key)
    if err != nil {
        b = &Blurb{Title: key}
    }
    user, _ := sessionUser(w, r)
    t := loadBlock("blurb")
    t.Execute(w, BlurbUser{Blurb:b, User:user})
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