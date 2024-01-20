package main

import (
    "net/http"
    "log"
    "html/template"
    "os"
    "encoding/json"
    "encoding/gob"
    "strconv"
    "bytes"
    // "fmt"

    "github.com/gorilla/sessions"
    "github.com/julienschmidt/httprouter"
)

var tmpl = map[string]*template.Template{}
var store *sessions.CookieStore
var accounts []*Account
var users []*User

type Account struct {
    Email string
    password []byte
    user int
}

func (a *Account) User() *User {
    if a.user < 0 {
        return nil
    }
    if len(users) < a.user {
        log.Printf("user for account [%s] not found.", a.Email)
        return nil
    }
    u := users[a.user]
    if u.Email == a.Email {
        return u
    }
    log.Printf("Emails for account: %s and associated user: %s do not match", a.Email, u.Email)
    return nil
}

func (a *Account) toMap() map[string]string {
    var aMap map[string]string
    aMap["Email"] = a.Email
    aMap["password"] = string(a.password)
    aMap["user"] = string(a.user)
    return aMap
}

func (a *Account) validatePassword(p string) bool {
    // I am fully aware this is all incorrect and incredibly insecure. This is still development phase
    return bytes.Equal(a.password, []byte(p))
}

type User struct {
    Email string
    Name string
    Admin bool
}

func (u *User) toMap() map[string]string {
    var uMap map[string]string
    uMap["Email"] = u.Email
    uMap["Name"] = u.Name
    uMap["Admin"] = "false"
    if u.Admin {
        uMap["Admin"] = "true"
    }
    return uMap
}

type Blurb struct {
    Title string
    Body []byte
}

type BlurbUser struct {
    Blurb *Blurb
    User *User
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
    user, yes := session.Values["user"].(User)
    // log.Printf("%s; %s", user, session.Values["user"])
    if yes {
        return &user, nil
    }
    return nil, nil
}

func sessionUserSet(w http.ResponseWriter, r *http.Request, user *User) {
    session, err := store.Get(r, "user-session")
    if err != nil {
        log.Fatal(err)
    }
    session.Values["user"] = user
    err = session.Save(r, w)
	if err != nil {
        log.Fatal(err)
	}
}

func LoginPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    acc := getAccount(r.PostFormValue("email"))
    if acc == nil {
        // w.Header().Set("HX-Redirect", "/")
        w.WriteHeader(http.StatusNotFound)
        return
    }
    if !acc.validatePassword(r.PostFormValue("password")) {
        // w.Header().Set("HX-Redirect", "/")
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    user := acc.User()
    sessionUserSet(w, r, user)
    w.Header().Set("HX-Redirect", "/")
    w.WriteHeader(http.StatusOK)
}

func getAccount(email string) *Account {
    for i := range accounts {
        if accounts[i].Email == email {
            return accounts[i]
        }
    }
    return nil
}

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

func loadAccounts() {
    accounts = nil
    data, err := loadJSON("Data/Users/accounts.json")
    if err != nil {
        log.Fatal("Loading account JSON failed")
        return
    }
    for a := range data {
        userKey, err := strconv.Atoi(data[a]["user"])
        if err != nil {
            userKey = -1
            log.Printf("No User found for Account: %s", data[a]["Email"])
        }
        account := &Account{Email: data[a]["Email"], password: []byte(data[a]["password"]), user: userKey}
        accounts = append(accounts, account)
    }
}

func loadUsers() {
    users = nil
    data, err := loadJSON("Data/Users/users.json")
    if err != nil {
        log.Fatal("Loading user JSON failed")
        return
    }
    for a := range data {
        admin := data[a]["Admin"] == "true"
        user := &User{Email: data[a]["Email"], Name: data[a]["Name"], Admin: admin}
        users = append(users, user)
    }
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

func LoginPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    t := loadPage("login")
    t.ExecuteTemplate(w, "base", nil)
}

func profilePage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    user, err := sessionUser(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    if user == nil {
        http.Redirect(w, r, "/users/login/", http.StatusNotFound)
        return
    }
    t := loadPage("profile")
    t.ExecuteTemplate(w, "base", user)
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
    router.GET("/users/profile", profilePage)
    router.GET("/blurb/:key/edit", BlurbEdit)
    router.POST("/blurb/:key/save", BlurbSave)
    router.GET("/blurb/:key", BlurbGet)

    router.ServeFiles("/static/*filepath", http.Dir("static"))

    log.Fatal(http.ListenAndServe(":8080", router))
}