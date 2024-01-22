package main

import (
    "net/http"
    "log"
    // "html/template"
    // "os"
    // "encoding/json"
    // "encoding/gob"
    "strconv"
    // "bytes"
    "fmt"
	
    "github.com/julienschmidt/httprouter"
)

type TeamMember struct {
    Name string
    Title string
    ShortDesc string
    Key int
}

// Team Members

func loadTeamMembers() []*TeamMember {
    var team []*TeamMember
    data, err := loadJSON("Data/theteam.json")
    if err != nil {
        log.Fatal("Loading team members JSON failed")
        return team
    }
    for i := range data {
        memb := &TeamMember{Title: data[i]["Title"], Name: data[i]["Name"], ShortDesc: data[i]["ShortDesc"], Key:i}
        team = append(team, memb)
    }
    return team
}

func MemberBlock(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    team := loadTeamMembers()
    i, err := strconv.Atoi(key)
    if err != nil || len(team) < i {
        m := &Message{Text:fmt.Sprintf("Team Member does not exist for ID: %s", key), Error:true}
        sendMessage(w, m, http.StatusNotFound)
        return 
    }
    memb := team[i]
    t := loadBlock("member")
    t.Execute(w, memb)
}

func TheTeam(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    t := loadPage("theteam")
    team := loadTeamMembers()
    t.ExecuteTemplate(w, "base", team)
}