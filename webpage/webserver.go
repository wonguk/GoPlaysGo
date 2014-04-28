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
)

type Page struct {
	Title string
	Body  []byte
}

// Allows pages to be saved to disk
func (p *Page) save() error {
	filename := "Ai/" + p.Title + ".go"
	return ioutil.WriteFile(filename,p.Body,0600)
}


// Allows loading pages saved to disk
func loadPage(title string) (*Page, error) {
    filename := "Ai/" + title + ".go"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

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

func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

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

func runHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/run/"):]
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    /* Enter what ever code we want to show here*/
    wg := new(sync.WaitGroup)
    wg.Add(1)
    go exe_cmd("test.bat",wg)
    wg.Wait()
    p.Body=([]byte("This is one long test of results \n Lets see if it printed a new line!"))

    renderTemplate(w, "run", p)
}

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


func main() {
	http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    http.HandleFunc("/run/",   runHandler)
    http.ListenAndServe(":8080", nil)
}

