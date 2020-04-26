package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"path/filepath"
	"server/movies"
	"sync"
)

func recursiveRead(dirPath string, wg *sync.WaitGroup, fileInfo func(string, string, *sync.WaitGroup, *sql.DB, *movies.ItemList, *movies.ItemList), db *sql.DB, mlist *movies.ItemList, slist *movies.ItemList)  {
	defer wg.Done()
	items, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error Reading Dir: ", err)
		return
	}
	for _, item := range items {
		wg.Add(1)
		path := filepath.Join(dirPath, item.Name())
		if item.IsDir() {
			go recursiveRead(path, wg, fileInfo, db, mlist, slist)
		} else {
			go fileInfo(item.Name(), path, wg, db, mlist, slist)
		}
	}
}

func scan(dirPath string, fileInfo func(string, string, *sync.WaitGroup, *sql.DB, *movies.ItemList, *movies.ItemList), db *sql.DB, mlist *movies.ItemList, slist *movies.ItemList)  {
	var wg sync.WaitGroup
	wg.Add(1)
	go recursiveRead(dirPath, &wg, fileInfo, db, mlist, slist)
	wg.Wait()
}

func main()  {
	database, err := sql.Open("sqlite3", "./temp.db")
	if err != nil {
		fmt.Println("Error Opening database", err)
		return
	}
	defer database.Close()
	movies.PrepareDb(database)
	movieList := movies.GetAllRecords(database, "movies")
	subtitleList := movies.GetAllRecords(database, "subtitles")
	testPath, err := filepath.Abs("test")
	if err != nil {
		fmt.Println("Error getting current path", err)
		return
	}
	scan(testPath, movies.ReadFileInfo, database, movieList, subtitleList)
	movies.CleanDb(database, movieList, "movies")
	movies.CleanDb(database, subtitleList, "subtitles")
}
