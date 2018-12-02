package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/zmb3/spotify"
)


var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

var tpl *template.Template
const redirectURI = "http://localhost:8080/callback"

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/home", home)
	http.HandleFunc("/analyze", analyzePlaylist)
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("images"))))
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	authUrl := auth.AuthURL(state)
	parsedUrl, err := url.Parse(authUrl)

	urlStruct := struct {
		Host string
		Parameters url.Values
	}{
		"https://" + parsedUrl.Hostname() + parsedUrl.Path,
		parsedUrl.Query(),
	}

	err = tpl.ExecuteTemplate(w, "index.gohtml", urlStruct)
	HandleError(w, err)
}

// TODO: Need to get all of the playlist by iterating
//.      through the pages returned
func home(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("Auth")
	HandleError(w, err)

	tok := GetAuthToken(c, w)

	client := auth.NewClient(tok)
	playlists, err := client.CurrentUsersPlaylists()
	HandleError(w, err)

	user, err := client.CurrentUser()
	HandleError(w, err)

	plInfo := struct {
		Name string
		ProfilePic string
		Playlists []spotify.SimplePlaylist
	} {
		user.User.DisplayName,
		user.User.Images[0].URL,
		playlists.Playlists,
	}

	err = tpl.ExecuteTemplate(w, "home.gohtml", plInfo)
	HandleError(w, err)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	SetAuthCookie(w, tok)

	http.Redirect(w, r, "/home", http.StatusFound)
}
