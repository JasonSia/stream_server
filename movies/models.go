package movies

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

type movie struct {
	id uuid.UUID
	name string
	fileName string
	path string
	year uint64
	width int64
	height int64
	status uint8
	duration time.Duration
	videoCodec string
	audioCodec string
	subtitles []subtitle
}

type subtitle struct {
	id uuid.UUID
	fileName string
	title string
	path string
}

const (
	UNWATCHED = iota
	PLANNED = iota
	WATCHED = iota
)

// In minutes
const MinMovieTime = 15

type ItemList struct {
	mu    sync.Mutex
	items map[string]uuid.UUID
}