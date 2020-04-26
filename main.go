package main

import (
	"fmt"
	mediainfo "github.com/lbryio/go_mediainfo"
	"io/ioutil"
	"path/filepath"
	"sync"
)

func readFileInfo(fileName, filePath string, wg *sync.WaitGroup)  {
	defer wg.Done()
	info, err := mediainfo.GetMediaInfo(filePath)
	if err != nil {
		fmt.Println("Error in reading: ", err)
	}
	if info.Video.CodecID != "" {
		fmt.Println("Video", info)
	} else if info.SubtitlesCnt != 0 {
		fmt.Println("Subtitles", info)
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
