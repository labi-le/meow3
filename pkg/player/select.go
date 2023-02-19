package player

import (
	"fmt"
	"io/fs"
	"math/rand"
)

func SelectRandomTrack(dir fs.FS) string {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		panic("failed to read dir: " + err.Error())
		return ""
	}

	if len(entries) == 0 {
		panic("error read dir")
	}

	// select random track
	track := entries[rand.Intn(len(entries))].Name()

	// return absolute path
	return fmt.Sprintf("%s/%s", dir, track)
}

type Playlist []Music

func (p Playlist) String() string {
	text := ""
	for _, track := range p {
		text += fmt.Sprintf("%s\n", track.Track)
	}

	return text
}

type Music struct {
	Track string
	Path  string
}

func GetAllTracks(dir fs.FS) Playlist {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		panic("failed to read dir: " + err.Error())
	}

	tracks := make(map[string]string, len(entries))
	playlist := make(Playlist, len(entries))
	for _, entry := range entries {
		tracks[entry.Name()] = fmt.Sprintf("%s/%s", dir, entry.Name())
		playlist = append(playlist, Music{
			Track: entry.Name(),
			Path:  fmt.Sprintf("%s/%s", dir, entry.Name()),
		})
	}

	return playlist
}
