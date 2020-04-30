package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"io/ioutil"
	"log"
	"path/filepath"
	"server/movies"
	"sync"
	"time"
)

func recursiveRead(dirPath string, wg *sync.WaitGroup, fileInfo func(string, string, *sync.WaitGroup, *pgxpool.Pool, *movies.ItemList, *movies.ItemList, *time.Time, *time.Time), db *pgxpool.Pool, mlist *movies.ItemList, slist *movies.ItemList, lastScan *time.Time)  {
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
			go recursiveRead(path, wg, fileInfo, db, mlist, slist, lastScan)
		} else {
			t := item.ModTime()
			go fileInfo(item.Name(), path, wg, db, mlist, slist, lastScan, &t)
		}
	}
}

func scan(dirPath string, fileInfo func(string, string, *sync.WaitGroup, *pgxpool.Pool, *movies.ItemList, *movies.ItemList, *time.Time, *time.Time), db *pgxpool.Pool, mlist *movies.ItemList, slist *movies.ItemList, lastScan *time.Time)  {
	var wg sync.WaitGroup
	wg.Add(1)
	go recursiveRead(dirPath, &wg, fileInfo, db, mlist, slist, lastScan)
	wg.Wait()
}

func main()  {
	start := time.Now()
	pool, err := setUpDb()
	if err != nil {
		return
	}
	defer pool.Close()
	movieList := movies.GetAllRecords(pool, "movies")
	subtitleList := movies.GetAllRecords(pool, "subtitles")
	testPath, err := filepath.Abs("/mnt/media/Videos/Hollywood Movies")
	if err != nil {
		fmt.Println("Error getting current path", err)
		return
	}
	t := time.Now()
	scan(testPath, movies.ReadFileInfo, pool, movieList, subtitleList, &t)
	movies.CleanDb(pool, movieList, "movies")
	movies.CleanDb(pool, subtitleList, "subtitles")
	elapsed := time.Since(start)
	log.Printf("Took %s", elapsed)
}

func setUpDb() (*pgxpool.Pool, error) {
	conConfig, err := pgxpool.ParseConfig("postgres://ayush:testpass@localhost/test")
	if err != nil {
		fmt.Println("Error Opening database", err)
		return nil, err
	}
	conConfig.MinConns = 5
	conConfig.MaxConns = 12
	pool, err := pgxpool.ConnectConfig(context.Background(), conConfig)
	if err != nil {
		fmt.Println("Error Opening database", err)
		return nil, err
	}
	movies.PrepareDb(pool)
	return pool, err
}
