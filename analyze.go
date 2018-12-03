package main

import (
	"net/http"
	"time"

	"github.com/zmb3/spotify"
	// "golang.org/x/oauth2"
)

type analysisData struct {
	tempo float64
}

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

		var pl spotify.SimplePlaylist
		var tracks []spotify.PlaylistTrack
		for i := range playlists.Playlists {
			if playlists.Playlists[i].Name == playlistName {
				tracks = getTracks(playlists.Playlists[i], client, w)
				pl = playlists.Playlists[i]
			}
		}

		analyzeTracks(pl, tracks, w, client)
	} else {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func getTracks(pl spotify.SimplePlaylist, client spotify.Client, w http.ResponseWriter) []spotify.PlaylistTrack {
	ctry := "US"
	opt := spotify.Options {Country: &ctry}
	t, err := client.GetPlaylistTracksOpt(pl.ID, &opt, "total")
	HandleError(w, err)
	total := t.Total

 	l := 20
	var tracks []spotify.PlaylistTrack
	for i := 0; i+20 <= total; i += 20 {
		opt = spotify.Options {
			Limit: &l,
			Offset: &i,
		}
		t, err = client.GetPlaylistTracksOpt(pl.ID, &opt, "")
		HandleError(w, err)

		tracks = append(tracks, t.Tracks...)
	}
	leftover := total % 20
	ofst := len(tracks)
	opt = spotify.Options {
		Limit: &leftover,
		Offset: &ofst,
	}

	t, err = client.GetPlaylistTracksOpt(pl.ID, &opt, "")
	HandleError(w, err)

	tracks = append(tracks, t.Tracks...)

	return tracks
}

func analyzeTracks(pl spotify.SimplePlaylist, tracks []spotify.PlaylistTrack, w http.ResponseWriter, client spotify.Client) {
	firstTrack := getFirstTrack(tracks, w)
	data := audioAnalysis(tracks, w, client)

	t, err := time.Parse(time.RFC3339, firstTrack.AddedAt)
	HandleError(w, err)
	firstTrack.AddedAt = t.Format(time.RFC822)

	a := struct {
		Playlist spotify.SimplePlaylist
		FirstTrack spotify.PlaylistTrack
		Tempo, BPS float64
	} {
		pl,
		firstTrack,
		data.tempo,
		60.0 / data.tempo,
	}

	err = tpl.ExecuteTemplate(w, "analyze.gohtml", a)
	HandleError(w, err)

}

func getFirstTrack(tracks []spotify.PlaylistTrack, w http.ResponseWriter) spotify.PlaylistTrack {
	earliestTime := time.Unix(1<<63-62135596801, 999999999)
	var firstTrack spotify.PlaylistTrack
	for i := range tracks {
		t, err := time.Parse(time.RFC3339, tracks[i].AddedAt)
		HandleError(w, err)
		if t.Before(earliestTime) {
			earliestTime = t
			firstTrack = tracks[i]
		}
	}
	return firstTrack
}

func audioAnalysis(tracks []spotify.PlaylistTrack, w http.ResponseWriter, client spotify.Client) analysisData {
	var aa []*spotify.AudioAnalysis

	for i := range tracks {
		a, err := client.GetAudioAnalysis(tracks[i].Track.SimpleTrack.ID)
		aa = append(aa, a)
		HandleError(w, err)
	}

	var tempo float64
	for i := range aa {
		tempo += aa[i].Track.Tempo
	}

	tempo = tempo / float64(len(tracks))

	data := analysisData {
		tempo,
	}

	return data

}

