package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = template.Must(template.ParseFiles("src/edit.html", "src/view.html"))

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view.html", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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

func Sock(ws *websocket.Conn) {
	var err error
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}
		msg := "Received:  " + reply
		if reply == "give me tasks" {
			msg = `task of the day, code!!`
			if err = websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("Can't send")
				break
			}
		} else {
			fmt.Println("Received back from client: " + reply)

			fmt.Println("Sending to client: " + msg)

			if err = websocket.Message.Send(ws, msg); err != nil {
				fmt.Println("Can't send")
				break
			}
		}
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", file)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/", staticHandler)
	http.Handle("/ws", websocket.Handler(Sock))
	http.ListenAndServe(":8080", nil)
}
