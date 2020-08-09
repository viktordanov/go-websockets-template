package model

import (
	"errors"
	"sync"
)

type Room struct {
	Owner        string
	Name         string
	Playlist     []string
	CurrentIndex int
	Users        map[string]*User
	mux          *sync.Mutex
}

func NewRoom(owner string, name string, currentURL string) *Room {
	return &Room{
		Owner:        owner,
		Name:         name,
		Playlist:     []string{},
		CurrentIndex: 0,
		Users:        map[string]*User{},
		mux:          &sync.Mutex{},
	}
}

func (r *Room) UserCount() int {
	return len(r.Users)
}

func (r *Room) NextOwner() {
	for _, user := range r.Users {
		if user.Name != r.Owner {
			r.Owner = user.Name
			break
		}
	}
}

func (r *Room) CurrentUrl() string {
	if r.CurrentIndex > len(r.Playlist)-1 {
		return ""
	}
	return r.Playlist[r.CurrentIndex]
}

func (r *Room) AddToPlaylist(url string) error {
	r.mux.Lock()

	r.Playlist = append(r.Playlist, url)
	r.mux.Unlock()
	return nil
}

func (r *Room) RemoveFromPlaylist(url string) error {
	r.mux.Lock()

	urlAtIndex := r.CurrentUrl()

	index := -1
	for i, u := range r.Playlist {
		if u == url {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("Video not found")
	}
	copy(r.Playlist[index:], r.Playlist[index+1:]) // Shift r.Playlist[i+1:] left one index.
	r.Playlist[len(r.Playlist)-1] = ""             // Erase last element (write zero value).
	r.Playlist = r.Playlist[:len(r.Playlist)-1]    // Truncate slice.

	for i, s := range r.Playlist {
		if s == urlAtIndex {
			r.SetCurrentIndex(i)
			break
		}
	}

	r.mux.Unlock()
	return nil
}

func (r *Room) NextUrl() string {
	if r.CurrentIndex+1 >= len(r.Playlist) {
		r.CurrentIndex = 0
	} else {
		r.CurrentIndex++
	}
	return r.CurrentUrl()
}
func (r *Room) PrevUrl() string {
	if r.CurrentIndex-1 < 0 {
		r.CurrentIndex = len(r.Playlist) - 1
	} else {
		r.CurrentIndex--
	}
	return r.CurrentUrl()
}

func (r *Room) GoToFirst() {
	r.SetCurrentIndex(0)
}

func (r *Room) GoToLast() {
	r.SetCurrentIndex(len(r.Playlist) - 1)
}

func (r *Room) GetCurrentIndex() int {
	return r.CurrentIndex
}
func (r *Room) SetCurrentIndex(newIndex int) {
	r.CurrentIndex = newIndex
}
func (r *Room) GetPlaylist() []string {
	return r.Playlist
}

func (r *Room) UserNames() []string {
	names := []string{}
	for name := range r.Users {
		names = append(names, name)
	}
	return names
}
