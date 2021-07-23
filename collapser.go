package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type ControlFile struct {
	sync.Mutex
	files      []string
	subdir     string
	timestamp  int64
	is_working bool
	is_done    bool
}

func (v *ControlFile) DebugToStr() string {

	str := ""
	files := ""
	for _, fn := range v.files {
		files = files + fn + " "
	}

	str = str + fmt.Sprintf("timestamp %d %s with files %s (is-working %v is-done %v)\n", v.timestamp, v.subdir, files, v.is_working, v.is_done)

	return str
}

type ControlFileManager struct {
	sync.Mutex
	jobs []*ControlFile
}

func (cm *ControlFileManager) AddControlFile(frl *FileResponseList) {

	job := ControlFile{
		timestamp: frl.GetTimestamp(),
		files:     frl.GetFilesNames(),
		subdir:    frl.GetSubdir(),
	}
	cm.Lock()
	cm.jobs = append(cm.jobs, &job)
	cm.Unlock()
}

func (cm *ControlFileManager) AddControlFromDir(fromDir string) error {
	cm.Lock()
	defer cm.Unlock()

	println("Scanning " + fromDir + " *.control")
	err := filepath.WalkDir(fromDir, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(file) != ".control" {
			return nil
		}
		println("Found file " + file)

		controlFileEpoch := strings.Split(path.Base(file), ".")

		timestamp, err := strconv.Atoi(controlFileEpoch[0])
		if err != nil {
			return err
		}

		subdir := path.Base(path.Dir(file))

		fmt.Printf(" timestamp %d\n", timestamp)
		println(" subdir " + subdir)

		readFile, err := os.Open(file)

		if err != nil {
			return err
		}
		defer readFile.Close()

		var referencedFiles []string
		scanner := bufio.NewScanner(readFile)

		for scanner.Scan() {
			referencedFiles = append(referencedFiles, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		for _, v := range referencedFiles {
			println(" - " + v)
		}

		job := &ControlFile{
			timestamp: int64(timestamp),
			files:     referencedFiles,
			subdir:    subdir,
		}
		cm.jobs = append(cm.jobs, job)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (cm *ControlFileManager) DebugToStr() string {

	str := ""
	cm.Lock()
	defer cm.Unlock()
	for i, v := range cm.jobs {

		str = str + fmt.Sprintf("entry %d, %s", i, v.DebugToStr())
	}

	return str
}

func NewControlFileManager() *ControlFileManager {
	return &ControlFileManager{
		jobs: make([]*ControlFile, 0, 100),
	}
}
