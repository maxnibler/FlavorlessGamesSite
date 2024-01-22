package main

import (
    "net/http"
    "log"
    // "html/template"
    // "os"
    // "encoding/json"
    // "encoding/gob"
    "strconv"
    "bytes"
    "fmt"
	
    "github.com/julienschmidt/httprouter"
)

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

func addAccount(acc *Account) {
    accounts = append(accounts, acc)
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

func addUser(user *User) {
    users = append(users, user)
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
    // t := loadBlock("login")
    acc := getAccount(r.PostFormValue("email"))
    if acc == nil {
        m := &Message{Text:fmt.Sprintf("No Account for the email: '%s' could be found", r.PostFormValue("email")),Error:true}
        sendMessage(w, m, http.StatusNotFound)
        return
    }
    if !acc.validatePassword(r.PostFormValue("password")) {
        m := &Message{Text:fmt.Sprintf("Password entered does not match the account: %s", r.PostFormValue("email")),Error:true}
        sendMessage(w, m, http.StatusUnauthorized)
        return
    }
    user := acc.User()
    sessionUserSet(w, r, user)
    w.Header().Set("HX-Redirect", "/users/profile")
    w.WriteHeader(http.StatusOK)
}

func SignupPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // t := loadBlock("signup")
    acc := getAccount(r.PostFormValue("email"))
    if acc != nil {
        m := &Message{Text:fmt.Sprintf("Account for the email: '%s' already exists", r.PostFormValue("email")),Error:true}
        sendMessage(w, m, http.StatusConflict)
        return
    }
    if r.PostFormValue("password") != r.PostFormValue("password-verify") {
        m := &Message{Text:"Passwords do not match",Error:true}
        sendMessage(w, m, http.StatusNotAcceptable)
        return
    }

    user := &User{Name:r.PostFormValue("username"), Email:r.PostFormValue("email"), Admin:false}
    i := len(users)
    addUser(user)
    acc = &Account{Email:r.PostFormValue("email"), password:[]byte(r.PostFormValue("password")), user:i}
    addAccount(acc)
    sessionUserSet(w, r, user)
    w.Header().Set("HX-Redirect", "/users/profile")
    w.WriteHeader(http.StatusOK)
}

func SignupForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    t := loadBlock("signup")
    t.Execute(w, nil)
}

func LoginForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    t := loadBlock("login")
    t.Execute(w, nil)
}

func getAccount(email string) *Account {
    for i := range accounts {
        if accounts[i].Email == email {
            return accounts[i]
        }
    }
    return nil
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