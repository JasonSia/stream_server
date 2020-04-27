package movies

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	mediainfo "github.com/lbryio/go_mediainfo"
	"log"
	"sync"
	"time"
)

func ReadFileInfo(fileName, filePath string, wg *sync.WaitGroup, db *pgxpool.Pool, mlist *ItemList, slist *ItemList)  {
	defer wg.Done()
	info, err := mediainfo.GetMediaInfo(filePath)
	if err != nil {
		fmt.Println("Error in reading: ", err)
	}
	if info.Video.CodecID != "" {
		wg.Add(1)
		go processMovies(fileName, filePath, wg, db, mlist, info)
	} else if info.SubtitlesCnt != 0 {
		wg.Add(1)
		go processSubtitles(fileName, filePath, wg, db, info, slist)
	}
}

func processMovies(fileName, filePath string, wg *sync.WaitGroup, db *pgxpool.Pool, mlist *ItemList, info *mediainfo.SimpleMediaInfo)  {
	defer wg.Done()
	el, ok := mlist.items[filePath]
	if ok{
		m := movie{
			id: el,
			width: info.Video.Width,
			height: info.Video.Height,
			duration: time.Duration(info.General.Duration),
			videoCodec: info.Video.CodecID,
			audioCodec: info.Audio.CodecID,
		}
		m.Update(db)
		mlist.mu.Lock()
		delete(mlist.items, filePath)
		mlist.mu.Unlock()
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			fmt.Println("Error in creating id", err)
			return
		}
		m := movie{
			id: id,
			fileName: fileName,
			name: fileName,
			path: filePath,
			year: 2010,
			width: info.Video.Width,
			height: info.Video.Height,
			status: UNWATCHED,
			duration: time.Duration(info.General.Duration),
			videoCodec: info.Video.CodecID,
			audioCodec: info.Audio.CodecID,
			subtitles: make([]subtitle, 0),
		}
		m.Add(db)
	}
}

func processSubtitles(fileName, filePath string, wg *sync.WaitGroup, db *pgxpool.Pool, info *mediainfo.SimpleMediaInfo, slist *ItemList)  {
	defer wg.Done()
	el, ok := slist.items[filePath]
	if ok {
		s := subtitle{
			id: el,
			title: info.Subtitles[0].Title,
		}
		s.Update(db)
		slist.mu.Lock()
		delete(slist.items, filePath)
		slist.mu.Unlock()
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			fmt.Println("Error in creating id", err)
			return
		}
		s := subtitle{
			id: id,
			fileName: fileName,
			path: filePath,
			title: info.Subtitles[0].Title,
		}
		s.Add(db)
	}
}

func PrepareDb(db *pgxpool.Pool)  {
	_, err := db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS movies (id UUID PRIMARY KEY, name TEXT, fileName TEXT, path TEXT, year INTEGER, width INTEGER, height INTEGER, status INTEGER, duration INTEGER, video_codec TEXT, audio_codec TEXT)")
	if err != nil {
		fmt.Println("Error Creating Movies Table", err)
		return
	}
	_, err = db.Exec(context.Background(), "CREATE TABLE IF NOT EXISTS subtitles (id UUID PRIMARY KEY, fileName TEXT, path TEXT, title TEXT)")
	if err != nil {
		fmt.Println("Error Creating Subtitles Table", err)
		return
	}
}

func CleanDb(db *pgxpool.Pool, mlist *ItemList, tableName string)  {
	for _, val := range mlist.items {
		RemoveItem(db, val, tableName)
	}
}

func GetAllRecords(db *pgxpool.Pool, tableName string) *ItemList {
	var count uint
	err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM " + tableName).Scan(&count)
	rows, err := db.Query(context.Background(), "SELECT id, path FROM " + tableName)
	if err != nil {
		fmt.Println("Error in querying records", err)
		return &ItemList{
			items: make(map[string]uuid.UUID, 0),
		}
	}
	defer rows.Close()
	tempList := make(map[string]uuid.UUID, count)
	var path string
	var id uuid.UUID
	for rows.Next() {
		err := rows.Scan(&id, &path)
		if err != nil {
			log.Fatal(err)
		}
		tempList[path] = id
	}
	return &ItemList{
		items: tempList,
	}
}

func(m *movie) Add(db *pgxpool.Pool)  {
	 _, err := db.Exec(context.Background(), "INSERT INTO movies VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", m.id, m.name, m.fileName, m.path, m.year, m.width, m.height, m.status, m.duration, m.videoCodec, m.audioCodec)
	if err != nil {
		fmt.Println("Error Adding to database", err)
		return
	}
}

func(m *movie) Update(db *pgxpool.Pool)  {
	_, err := db.Exec(context.Background(), "UPDATE movies SET width = $2, height = $3, duration = $4, video_codec = $5, audio_codec = $6 WHERE id = $1", m.id, m.width, m.height, m.duration, m.videoCodec, m.audioCodec)
	if err != nil {
		fmt.Println("Error Updating database", err)
		return
	}
}

func(s *subtitle) Add(db *pgxpool.Pool)  {
	_, err := db.Exec(context.Background(), "INSERT INTO subtitles VALUES ($1, $2, $3, $4)", s.id, s.fileName, s.path, s.title)
	if err != nil {
		fmt.Println("Error Adding to database", err)
		return
	}
}

func(s *subtitle) Update(db *pgxpool.Pool)  {
	_, err := db.Exec(context.Background(), "UPDATE subtitles SET title = $2 WHERE id = $1", s.id, s.title)
	if err != nil {
		fmt.Println("Error Updating database", err)
		return
	}
}


func RemoveItem(db *pgxpool.Pool, id uuid.UUID, tableName string)  {
	_, err := db.Exec(context.Background(), "DELETE FROM " + tableName + " WHERE id = ?", id)
	if err != nil {
		fmt.Println("Error Removing movie", err)
		return
	}
}