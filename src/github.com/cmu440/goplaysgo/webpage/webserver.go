//WebServer.go :
//    Manages the html front end that calls and compiles the go code the servers are running
//    - Modifed code from go's html library and webapp tutorial.
package main

import (
    "bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	//"log"
	"os/exec"
	"strings"
	"sync"
    "github.com/cmu440/goplaysgo/rpc/mainrpc"
    "net/rpc"
    "strconv"
)
 

type Page struct {
	Title string
	Body  []byte
}

// save() -Allows pages to be saved to disk for future loading
func (p *Page) save() error {
	filename := "Ai/" + p.Title + ".go"
	return ioutil.WriteFile(filename,p.Body,0600)
}


//loadPage() Allows loading pages saved to disk
func loadPage(title string) (*Page, error) {
    filename := "Ai/" + title + ".go"
    body, err := ioutil.ReadFile(filename)
    fmt.Println(body,len(body))
    if len(body) == 2 {
        body = []byte("package ai \n \n" + "import" + " github.com/cmu440/goplaysgo/gogame \n \n" + "func NextMove(board gogame.Board, player gogame.Player) gogame.Move {...}" )
    }
    fmt.Println(body)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

//renderTemplate() -Draws the html page based on html request
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

//editHandler() -Manages the edit page where the user writes up the code
func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

//viewHandler() -Manages the view page where users see if their code compiled and can enter the test
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, err := loadPage(title)    
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }

    fmt.Println("Compiling:",title+".go")
    build := exec.Command("go", "build", "Ai/"+title+".go")
    var out bytes.Buffer
    build.Stderr = &out
    err = build.Run()
    if err != nil || len(out.String()) != 0 {
        p.Body= []byte("Error in Compling \n" + err.Error())
    } else {
        p.Body = []byte("Compling Passed run to test or edit")
    }
    renderTemplate(w, "view", p)
}

//saveHandler() -Manages saving the code written in editHandler to disk
func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//runHandler() -Runs the code and returns the results
func runHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/run/"):]
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    /* Enter what ever code we want to show here*/
    ai := new(mainrpc.SubmitAIArgs)
    ai.Name = title
    ai.Code,err = ioutil.ReadFile("Ai/" + title + ".go") 
    fmt.Println(ai)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("Error 4:",err)
        return
    }
    aiReply := new(mainrpc.SubmitAIReply)
    client,err :=  rpc.DialHTTP("tcp","71.199.115.110:9099")
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("Server offline")
        return
    }
    err = client.Call("MainServer.SubmitAI",ai,aiReply)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("SubmitAI error")
        return
    }
    if aiReply.Status != mainrpc.OK { //Make more cases later
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("Error 3:",err,aiReply.Status)
        return
    } 
    Standings := new(mainrpc.GetStandingsArgs)
    StandingsReply := new(mainrpc.GetStandingsReply)
    client,err = rpc.DialHTTP("tcp","71.199.115.110:9099")
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("client call error")
        return
    }
    err = client.Call("MainServer.GetStandings",Standings,StandingsReply)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("Getting Standings error")
        fmt.Println("Error 2:",err)
        return
    }
    if StandingsReply.Status != mainrpc.OK { //Make more cases later
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        fmt.Println("Error 1:",err)
        return
    } 
    var PageResults string = ""
    ServerStandings := StandingsReply.Standings
    for AI := 0; AI < len(ServerStandings); AI++ {
        PageResults = PageResults + "AI: "+ ServerStandings[AI].Name +" got: \n"
        PageResults = PageResults + "Wins: "+ strconv.Itoa(ServerStandings[AI].Wins) + "\n"
        PageResults = PageResults + "Losses: "+ strconv.Itoa(ServerStandings[AI].Losses) + "\n"
        PageResults = PageResults + "Draws: "+ strconv.Itoa(ServerStandings[AI].Draws) + "\n \n"
    }
    p.Body=([]byte(PageResults))

    renderTemplate(w, "run", p)
}

//exe_cmd() - A helper function that manages running command line instructions
// - Found in various tutorials
func exe_cmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println("Command is ",cmd)
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head,parts...).Output()
	if err != nil {
		fmt.Printf("%s",err)
	}
	fmt.Printf("%s",out)
	wg.Done()
}

//main() - Handles first visit and sets up ports and pages for user
func main() {
	http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    http.HandleFunc("/run/",   runHandler)
    http.ListenAndServe(":8080", nil)
}

