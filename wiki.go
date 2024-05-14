package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Page struct{
    Title string
    Body []byte
}

var templates = template.Must(template.ParseFiles("./pages/add.html", "./pages/edit.html", "./pages/view.html"))
var validPath = regexp.MustCompile("^/(edit|view|save|add)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    file, err := os.OpenFile(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0600)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = file.Write(p.Body)
    if err != nil {
        return err
    }
    return nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w,r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := os.ReadFile(filename)
    if(err != nil) {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func rederTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    title, err := getTitle(w,r)
    p, err := loadPage(title)
    if err != nil {
        log.Printf("Error loading page: %v", err)
        http.Redirect(w, r, "/edit/"+title, http.StatusNotFound)
        return
    }
    rederTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    title, err := getTitle(w,r)
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
        log.Printf("Error loading page: %v", err)
        http.Error(w, "Page not found", http.StatusNotFound)
        return
    }
    rederTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    title, err := getTitle(w,r)
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err = p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w , r, "/view/"+title, http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
    method := r.Method
    log.Printf("Method: %s", method)
    if method == "GET" {
        rederTemplate(w, "add", nil)
    } else {
        title := r.FormValue("title")
        body := r.FormValue("body")
        p := &Page{Title: title, Body: []byte(body)}
        err := p.save()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        http.Redirect(w , r, "/view/"+title, http.StatusFound)
    }
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/add/", addHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}


