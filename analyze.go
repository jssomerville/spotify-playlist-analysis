package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
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

// ~800 ms for 36 songs
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
	if total % 20 != 0 {
		leftover := total % 20
		ofst := len(tracks)
		opt = spotify.Options {
			Limit: &leftover,
			Offset: &ofst,
		}

		t, err = client.GetPlaylistTracksOpt(pl.ID, &opt, "")
		HandleError(w, err)

		tracks = append(tracks, t.Tracks...)
	}

	return tracks
}

func analyzeTracks(pl spotify.SimplePlaylist, tracks []spotify.PlaylistTrack, w http.ResponseWriter, client spotify.Client) {
	firstTrack := getFirstTrack(tracks, w)
	audioAnalysis(tracks, w, client)

	t, err := time.Parse(time.RFC3339, firstTrack.AddedAt)
	HandleError(w, err)
	firstTrack.AddedAt = t.Format(time.RFC822)

	// err = tpl.ExecuteTemplate(w, "analyze.gohtml", a)
	HandleError(w, err)
}

// ~ 27 micro sec for 36 tracks
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

func audioAnalysis(tracks []spotify.PlaylistTrack, w http.ResponseWriter, client spotify.Client) {
	trackChunks := ChunkSlice(tracks)
	var audioFeatures []*spotify.AudioFeatures
	for i := range trackChunks {
		af, err := client.GetAudioFeatures(trackChunks[i]...)
		HandleError(w, err)
		audioFeatures = append(audioFeatures, af...)
	}
	ch1 := make(chan float32)
	ch2 := make(chan float32)
	ch3 := make(chan int)
	ch4 := make(chan float32)
	ch5 := make(chan float32)

	go getAvgTempo(audioFeatures, ch1)
	go getAvgDance(audioFeatures, ch2)
	go getAvgDuration(audioFeatures, ch3)
	go getAvgEnergy(audioFeatures, ch4)
	go getAvgValence(audioFeatures, ch5)

	avgTempo, avgDance, avgDuration, avgEnergy, avgValence := <-ch1, <-ch2, <-ch3, <-ch4, <-ch5

	audioAnalysisData := struct {
		Tempo, Dance, Energy, Valence float32
		Duration int
	} {
		avgTempo,
		avgDance,
		avgEnergy,
		avgValence,
		avgDuration,
	}
}

func getAvgTempo(f []*spotify.AudioFeatures, c chan float32) {
	var avg float32
	for _, i := range f {
		avg += i.Tempo
	}
	avg /= float32(len(f))
	c <- avg
}

func getAvgDance(f []*spotify.AudioFeatures, c chan float32) {
	var avg float32
	for _, i := range f {
		avg += i.Danceability
	}
	avg /= float32(len(f))
	c <- avg
}

func getAvgDuration(f []*spotify.AudioFeatures, c chan int) {
	var avg int
	for _, i := range f {
		avg += i.Duration
	}
	avg /= len(f)
	c <- avg
}

func getAvgEnergy(f []*spotify.AudioFeatures, c chan float32) {
	var avg float32
	for _, i := range f {
		avg += i.Energy
	}
	avg /= float32(len(f))
	c <- avg
}

func getAvgValence(f []*spotify.AudioFeatures, c chan float32) {
	var avg float32
	for _, i := range f {
		avg += i.Valence
	}
	avg /= float32(len(f))
	c <- avg
}