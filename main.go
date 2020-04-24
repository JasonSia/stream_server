package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

var videoFormats = []string{"mp4", "mkv"}

func readFileInfo(fileName, filePath string, wg *sync.WaitGroup)  {
	defer wg.Done()
	for _, ext := range videoFormats {
		stat, _ := filepath.Match("*." + ext, fileName)
		if stat {
			fmt.Println(fileName)
		}
	}
}

func recursiveRead(dirPath string, wg *sync.WaitGroup)  {
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
			go recursiveRead(path, wg)
		} else {
			go readFileInfo(item.Name(), path, wg)
		}
	}
}

func scan(dirPath string)  {
	var wg sync.WaitGroup
	wg.Add(1)
	go recursiveRead(dirPath, &wg)
	wg.Wait()
}

func main()  {
	testPath, err := filepath.Abs("test")
	if err != nil {
		fmt.Println("Error getting current path", err)
	}
	scan(testPath)
}
