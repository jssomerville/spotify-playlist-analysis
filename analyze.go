package main

import (
"fmt"
	//"html/template"
	"net/http"

	// "github.com/zmb3/spotify"
	// "golang.org/x/oauth2"
)

func analyzePlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		c, err := r.Cookie("Auth")
		HandleError(w, err)

		tok := GetAuthToken(c, w)
		client := auth.NewClient(tok)

		r.ParseForm()
		playlistName := r.Form["playlistName"][0]

		playlists, err := client.CurrentUsersPlaylists()
		HandleError(w, err)

		for i := range playlists.Playlists {
			if playlists.Playlists[i].Name == playlistName {
				fmt.Println("Gotem")
			}
		}
	} else {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

