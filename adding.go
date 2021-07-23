package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
)

var (
	WorkDir = os.Getenv("WORK_DIR")
	BaseDir = os.Getenv("BASE_DIR")
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type FileResponse struct {
	FileID      string `json:"file_id"`
	OriginalRef string `json:"original_ref"`
}

func (fr *FileResponse) MoveFile(dest string) error {
	from := BaseDir + fr.OriginalRef
	fmt.Printf("moving %s to %s\n", from, dest)
	err := os.Rename(from, dest)
	if err != nil {
		return err
	}

	return nil
}

type FileResponseList struct {
	Files       []FileResponse `json:"files"`
	subdir      string
	timestamp   int64
	controlPath string
}

func (frl *FileResponseList) Add(file_name string) {

	added := FileResponse{
		FileID:      ksuid.New().String(),
		OriginalRef: file_name,
	}
	frl.Files = append(frl.Files, added)

}

func (frl *FileResponseList) GetFilesNames() []string {
	files := make([]string, 0, len(frl.Files))

	for _, fl := range frl.Files {
		files = append(files, fl.FileID)
	}

	return files
}

func (frl *FileResponseList) writeControlFile() error {

	if _, err := os.Stat(WorkDir + frl.subdir); os.IsNotExist(err) {
		err := os.MkdirAll(WorkDir+frl.subdir, 0775)
		if err != nil {
			return err
		}
	}

	logContent := ""
	fileContent := ""
	for _, fl := range frl.Files {
		fileContent = fileContent + fl.FileID + "\n"
		logContent = logContent + fl.FileID + " "
	}

	tmpName := fmt.Sprintf("%s/%d.tmp", WorkDir+frl.subdir, frl.timestamp)
	frl.controlPath = fmt.Sprintf("%s/%d.control", WorkDir+frl.subdir, frl.timestamp)

	fmt.Printf("writing control file %s for files %s\n", frl.controlPath, logContent)

	err := ioutil.WriteFile(tmpName, []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	err = os.Rename(tmpName, frl.controlPath)
	if err != nil {
		return err
	}

	return nil
}

func (frl *FileResponseList) processAll() error {

	if _, err := os.Stat(WorkDir + frl.subdir); os.IsNotExist(err) {
		err := os.MkdirAll(WorkDir+frl.subdir, 0775)
		if err != nil {
			return err
		}
	}

	// write first the control file
	// if the write or move fails, the files are still not moved, so client can retry safely
	// if this write succeed, probably the next move command will also succeed.
	// if program crash after move but before adding to the in memory queue,
	// control file will be scanned during next boot and processed
	err := frl.writeControlFile()
	if err != nil {
		return err
	}

	// TODO: deal with restoring moved files back if an error occurred
	for _, fl := range frl.Files {
		dest := WorkDir + frl.subdir + "/" + fl.FileID

		if err := fl.MoveFile(dest); err != nil {
			return err
		}
	}

	return nil
}

func newFileResponseList(cameraId int, timestamp int64) *FileResponseList {
	return &FileResponseList{
		Files:     make([]FileResponse, 0, 2),
		subdir:    fmt.Sprintf("cam-%d", cameraId),
		timestamp: timestamp,
	}
}

func adding_post(c echo.Context, onSuccess func(*FileResponseList) error) error {

	if onSuccess == nil {
		panic("you using it wrong! onSuccess must be defined")
	}

	cameraId, err := strconv.Atoi(c.QueryParam("camera_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			fmt.Sprintf("QueryParam camera_id invalid %s", err.Error()),
		})
	}
	timestamp, err := strconv.Atoi(c.QueryParam("timestamp"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			fmt.Sprintf("QueryParam timestamp invalid %s", err.Error()),
		})
	}

	names := c.QueryParams()["name"]
	file_list := newFileResponseList(cameraId, int64(timestamp))

	for _, filename := range names {
		if strings.Contains(filename, "../") {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				"File name cannot contain ../",
			})
		}

		fi, err := os.Stat(BaseDir + filename)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{err.Error()})
		} else if err != nil && os.IsNotExist(err) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				fmt.Sprintf("File %s does not exists", BaseDir+filename),
			})
		}
		if fi.Size() == 0 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				"File %s must have at least one byte",
			})
		}

		file_list.Add(filename)
	}

	if err := file_list.processAll(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			fmt.Sprintf("Cannot process: %s", err.Error()),
		})
	}

	if err := onSuccess(file_list); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			fmt.Sprintf("Cannot process: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, file_list)
}
