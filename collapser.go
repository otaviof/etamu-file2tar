package main

import (
	"fmt"
	"sync"
)

type ControlFile struct {
	files     []string
	timestamp int64
}

type ControlFileManager struct {
	sync.Mutex
	jobs []ControlFile
}

func (cm *ControlFileManager) AddControlFile(frl *FileResponseList) {

	job := ControlFile{
		timestamp: frl.timestamp,
		files:     frl.GetFilesNames(),
	}
	cm.Lock()
	cm.jobs = append(cm.jobs, job)
	cm.Unlock()
}

func (cm *ControlFileManager) DebugToStr() string {

	str := ""
	cm.Lock()
	defer cm.Unlock()
	for i, v := range cm.jobs {
		files := ""
		for _, fn := range v.files {
			files = files + fn + " "
		}

		str = str + fmt.Sprintf("entry %d - timestamp %d with files %s\n", i, v.timestamp, files)
	}

	return str
}

func NewControlFileManager() *ControlFileManager {
	return &ControlFileManager{
		jobs: make([]ControlFile, 0, 100),
	}
}
