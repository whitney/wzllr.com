package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

type NestedPage struct {
	HeadHTML   template.HTML
	NavHTML    template.HTML
	FooterHTML template.HTML
}

type Header struct {
	Title        string
}

type Nav struct {
	HomeClass    string
	AboutClass   string
	BikesClass   string
	GalleryClass string
}

type templateParser struct{
    HTML  string
}

func (tP *templateParser) Write(p []byte) (n int, err error){
    tP.HTML += string(p)
    return len(p), nil
}

func parseTemplate(filename string, data interface{}) (string){
	tp := &templateParser{}
	t, err := template.ParseFiles(filename)
	if err != nil {
		fmt.Printf("parseTemplate error: " + err.Error() + "\n")
	}

	t.Execute(tp, data)
	return tp.HTML
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

func loadNestedPage(title string) (*NestedPage, error) {
	// instantiate header template
	tHead := &Header{
		Title:title,
	}

	// instantiate nav template
	homeClass := ""
	aboutClass := ""
	bikesClass := ""
	galleryClass := ""
	switch title {
	case "home": homeClass = "active"
	case "about": aboutClass = "active"
	case "bikes": bikesClass = "active"
	case "gallery": galleryClass = "active"
	}
	tNav := &Nav{
		HomeClass:homeClass,
		AboutClass:aboutClass,
		BikesClass:bikesClass,
		GalleryClass:galleryClass,
	}

	// parse static footer
	footerHTML, err := ioutil.ReadFile("tmpl/footer.html")
	if err != nil {
		return nil, err
	}

	// instantiate NestedPage template
	tPage := &NestedPage{}
	tPage.HeadHTML   = template.HTML(parseTemplate("tmpl/header.html", tHead))
	tPage.NavHTML    = template.HTML(parseTemplate("tmpl/nav.html", tNav))
	tPage.FooterHTML = template.HTML(string(footerHTML))

	fmt.Printf("HeadHTML: " + "\n" + string(tPage.HeadHTML) + "\n")
	fmt.Printf("NavHTML: " + "\n" + string(tPage.NavHTML) + "\n")
	fmt.Printf("FooterHTML: " + "\n" + string(tPage.FooterHTML) + "\n")

	return tPage, nil
}

var templates = template.Must(template.ParseFiles("tmpl/about.html", "tmpl/home.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderNestedTemplate(w http.ResponseWriter, tmpl string, p *NestedPage) {
	fmt.Printf("\n\nHeaderHTML II: " + "\n" + string(p.HeadHTML) + "\n\n")
	fmt.Printf("\n\nFooterHTML II: " + "\n" + string(p.FooterHTML) + "\n\n")

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	/*
	t := template.New("Nested template")
	homeTmpl, err := ioutil.ReadFile("tmpl/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Printf(">>> homeTmpl: \n" + string(homeTmpl) + "\n")
	t, err = t.Parse(string(homeTmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	*/
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		fmt.Printf("viewHandler error: " + err.Error() + "\n")
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, title, p)
}

func nestedViewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadNestedPage(title)
	if err != nil {
		fmt.Printf("nestedViewHandler error: " + err.Error() + "\n")
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderNestedTemplate(w, title, p)
}

/*
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		fmt.Printf("saveHandler error: " + err.Error() + "\n")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
*/

const lenPath = len("/view/")

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := "home"
		fmt.Printf("URL.Path: " + r.URL.Path + "\n")
		if len(r.URL.Path) > 1 {
			title = r.URL.Path[1:]
		}
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

func sourceHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func main() {
    fmt.Printf("Starting HTTP server, listening on port 8081...\n")
	http.HandleFunc("/favicon.ico", sourceHandler)
	http.HandleFunc("/public/", sourceHandler)
	//http.HandleFunc("/", makeHandler(viewHandler))
	http.HandleFunc("/", makeHandler(nestedViewHandler))
	err := http.ListenAndServe(":80", nil)
    if err != nil {
       fmt.Printf("ListenAndServe Error: " + err.Error() + "\n")
    }
}
