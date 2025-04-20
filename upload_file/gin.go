package upload_file

import (
	"errors"
	"fmt"
	"github.com/a-aslani/wotop/model/apperror"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const (
	ErrInvalidFileType apperror.ErrorType = "ER0001 invalid file type %s"
	ErrFileSizeExceeds apperror.ErrorType = "ER0002 file size exceeds the maximum limit of %d bytes"
	ErrMissingFile     apperror.ErrorType = "ER0003 missing file"
)

type Params struct {
	FieldName  string
	IsRequired bool
	Path       string
	MaxSize    int64
	Accept     []string
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
	default:
		return "", ErrInvalidFileType
	}

	filePath := fmt.Sprintf("%s/%s.%s", params.Path, uuid.NewString(), ext)
	if err = c.SaveUploadedFile(fileHeader, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}
