package main

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"

	/*	_ "github.com/mattn/go-sqlite3"*/
	"io/ioutil"
	"log"
	"path/filepath"
	"server/movies"
	"sync"
	"time"
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
	driverConfig := stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
		},
	}

	stdlib.RegisterDriverConfig(&driverConfig)
	start := time.Now()
	database, err := sql.Open("pgx", "postgres://ayush:testpass@localhost/test")
	if err != nil {
		fmt.Println("Error Opening database", err)
		return
	}
	defer database.Close()
	movies.PrepareDb(database)
	movieList := movies.GetAllRecords(database, "movies")
	subtitleList := movies.GetAllRecords(database, "subtitles")
	testPath, err := filepath.Abs("/mnt/media/Videos/Hollywood Movies")
	if err != nil {
		fmt.Println("Error getting current path", err)
		return
	}
	scan(testPath, movies.ReadFileInfo, database, movieList, subtitleList)
	movies.CleanDb(database, movieList, "movies")
	movies.CleanDb(database, subtitleList, "subtitles")
	elapsed := time.Since(start)
	log.Printf("Took %s", elapsed)
}
