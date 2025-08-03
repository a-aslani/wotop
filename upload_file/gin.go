package upload_file

import (
	"errors"
	"fmt"
	"github.com/a-aslani/wotop/model/apperror"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ErrInvalidFileType apperror.ErrorType = "ER0001 invalid file type %s"
	ErrFileSizeExceeds apperror.ErrorType = "ER0002 file size exceeds the maximum limit of %d bytes"
	ErrMissingFile     apperror.ErrorType = "ER0003 missing file"
)

type Params struct {
	FieldName     string
	IsRequired    bool
	Path          string
	MaxSize       int64
	Accept        []string
	TempPattern   *string
	TempDir       *string
	SaveFileInDir bool
}

type fileUploader struct {
	FilePath *string  `json:"file_path"`
	FileSize int64    `json:"file_size"`
	Temp     *os.File `json:"temp"`
}

func NewUploader() *fileUploader {
	return &fileUploader{}
}

func (f *fileUploader) Path() string {
	if f.FilePath == nil {
		return ""
	}
	return *f.FilePath
}

func (f *fileUploader) Size() int64 {
	return f.FileSize
}

func (f *fileUploader) TempFile() *os.File {
	return f.Temp
}

func (f *fileUploader) Close() {
	if f.Temp != nil {
		defer f.Temp.Close()
	}
}

func (f *fileUploader) Upload(c *gin.Context, params Params) error {

	fileHeader, err := c.FormFile(params.FieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if params.IsRequired {
				return ErrMissingFile
			}
			return nil
		}
		return err
	}

	if fileHeader.Size > params.MaxSize {
		return ErrFileSizeExceeds.Var(params.MaxSize)
	}

	mimeType := fileHeader.Header.Get("Content-Type")

	isAccept := false

	for _, a := range params.Accept {
		if a == mimeType {
			isAccept = true
			break
		}
	}

	if !isAccept {
		return ErrInvalidFileType.Var(mimeType)
	}

	var tmpFile *os.File

	if params.TempDir != nil && params.TempPattern != nil {

		if *params.TempDir != "" {
			if err = os.MkdirAll(*params.TempDir, 0755); err != nil {
				return err
			}
		}

		tmpFile, err = os.CreateTemp(*params.TempDir, *params.TempPattern)
		if err != nil {
			return err
		}

		src, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(tmpFile, src)
		if err != nil {
			return err
		}

		if _, err = tmpFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}

	var filePath *string

	if params.SaveFileInDir {

		ext := filepath.Ext(fileHeader.Filename)

		finalDir := params.Path
		if err := os.MkdirAll(finalDir, 0755); err != nil {
			return err
		}

		filename := fmt.Sprintf("%s/%s.%s", params.Path, uuid.NewString(), ext)
		finalPath := filepath.Join(finalDir, filename)
		if err := c.SaveUploadedFile(fileHeader, finalPath); err != nil {
			return err
		}
	}

	f.FilePath = filePath
	f.FileSize = fileHeader.Size
	f.Temp = tmpFile

	return nil
}

func Upload(c *gin.Context, params Params) (string, error) {

	fileHeader, err := c.FormFile(params.FieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if params.IsRequired {
				return "", ErrMissingFile
			}
			return "", nil
		}
		return "", err
	}

	if fileHeader.Size > params.MaxSize {
		return "", ErrFileSizeExceeds.Var(params.MaxSize)
	}

	mimeType := fileHeader.Header.Get("Content-Type")

	isAccept := false

	for _, a := range params.Accept {
		if a == mimeType {
			isAccept = true
			break
		}
	}

	if !isAccept {
		return "", ErrInvalidFileType.Var(mimeType)
	}

	ext, err := getExt(mimeType)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("%s/%s.%s", params.Path, uuid.NewString(), ext)
	if err = c.SaveUploadedFile(fileHeader, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func getExt(mimeType string) (string, error) {

	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "application/pdf":
		ext = "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		ext = "docx"
	case "application/msword":
		ext = "doc"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		ext = "xlsx"
	case "application/vnd.ms-excel":
		ext = "xls"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		ext = "pptx"
	case "application/vnd.ms-powerpoint":
		ext = "ppt"
	case "application/zip":
		ext = "zip"
	case "application/x-rar-compressed":
		ext = "rar"
	case "application/x-7z-compressed":
		ext = "7z"
	case "application/x-tar":
		ext = "tar"
	case "application/x-gzip":
		ext = "gz"
	case "application/x-bzip2":
		ext = "bz2"
	case "application/x-xz":
		ext = "xz"
	case "application/x-zip-compressed":
		ext = "zip"
	case "text/csv", "application/csv", "text/comma-separated-values":
		ext = "csv"
	case "application/json", "text/json":
		ext = "json"
	default:
		return "", ErrInvalidFileType
	}

	return ext, nil
}
