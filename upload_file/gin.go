package upload_file

import (
	"errors"
	"fmt"
	"github.com/a-aslani/wotop/model/apperror"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	ErrInvalidFileType apperror.ErrorType = "ER0001 invalid file type %s"
	ErrFileSizeExceeds apperror.ErrorType = "ER0002 file size exceeds the maximum limit of %d bytes"
	ErrMissingFile     apperror.ErrorType = "ER0003 missing file"
)

// dirPerm is the mode a missing upload directory is created with. It is applied
// on creation only: a directory that already exists keeps the mode its owner
// gave it (see saveUploadedFile).
const dirPerm os.FileMode = 0o755

type Params struct {
	FieldName     string
	IsRequired    bool
	Path          string
	MaxSize       int64
	Accept        []string
	TempPattern   *string
	TempDir       *string
	SaveFileInDir bool

	// Extensions maps a content type to the extension it is stored under,
	// taking precedence over the built-in table. It is the escape hatch for a
	// type the built-in table does not know, so a new one does not need a
	// release of this module. It does not widen what is accepted — Accept
	// alone decides that.
	Extensions map[string]string
}

type fileUploader struct {
	FilePath *string  `json:"file_path"`
	FileSize int64    `json:"file_size"`
	Temp     *os.File `json:"temp"`
	Ext      string   `json:"ext"`
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

	ext := filepath.Ext(fileHeader.Filename)
	f.Ext = strings.ToLower(ext)

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
			if err = os.MkdirAll(*params.TempDir, dirPerm); err != nil {
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

		// f.Ext already carries the leading dot, so it is appended rather than
		// joined with one, and the name is joined to the directory once. Both
		// were wrong before: the path was built from params.Path and then
		// joined to params.Path again, and the dot was doubled.
		finalPath := fmt.Sprintf("%s/%s%s", strings.TrimRight(params.Path, "/"), uuid.NewString(), f.Ext)

		if err = saveUploadedFile(fileHeader, finalPath); err != nil {
			return err
		}

		// Previously left nil, which made Path() report "" for a file that had
		// in fact been written.
		filePath = &finalPath
	}

	f.FilePath = filePath
	f.FileSize = fileHeader.Size
	f.Temp = tmpFile

	return nil
}

// Upload stores one file from the multipart form and returns the path it was
// written to. A field that was not sent returns "" unless it is required.
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

	ext, err := getExt(mimeType, params.Extensions)
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("%s/%s.%s", strings.TrimRight(params.Path, "/"), uuid.NewString(), ext)

	if err = saveUploadedFile(fileHeader, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

// saveUploadedFile writes the uploaded file to path, creating the directory if
// it is missing.
//
// Deliberately not gin's Context.SaveUploadedFile. Since gin v1.11 that method
// chmods the destination directory on every write:
//
//	if err = os.MkdirAll(dir, mode); err != nil { return err }
//	if err = os.Chmod(dir, mode); err != nil { return err }
//
// A process that shares an upload directory with another service — one volume,
// several services writing into it — does not own that directory, so the chmod
// comes back EPERM and fails the upload: "chmod uploads/x: operation not
// permitted". Nothing is wrong with the upload; the directory simply belongs to
// someone else.
//
// A directory that already exists is therefore left exactly as it is: whoever
// owns it decides its mode. Owning the write also makes this package behave the
// same whichever gin version a consumer's module graph resolves to, which is not
// obvious — this module requires v1.10, where no chmod happens, so a consumer on
// v1.11 gets behaviour this module's own tests would never see.
func saveUploadedFile(fileHeader *multipart.FileHeader, path string) error {

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return err
		}
	}

	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)

	return err
}

// extensions maps an accepted content type to the extension it is stored under.
//
// This is not the allow-list: Params.Accept decides what a handler takes, and a
// type is only looked up here once it has been accepted. A type missing from
// this table therefore surfaces as "invalid file type" for a file the handler
// meant to allow, which is why the image entries are kept broad and why
// Params.Extensions exists.
var extensions = map[string]string{
	// Images.
	"image/jpeg":    "jpg",
	"image/jpg":     "jpg", // not a registered type, but browsers and older clients send it
	"image/png":     "png",
	"image/svg+xml": "svg",
	"image/gif":     "gif",
	"image/webp":    "webp",
	"image/avif":    "avif",
	"image/bmp":     "bmp",
	"image/tiff":    "tiff",
	"image/x-icon":  "ico",
	"image/heic":    "heic",

	// Documents.
	"application/pdf":    "pdf",
	"application/msword": "doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
	"application/vnd.ms-excel": "xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "xlsx",
	"application/vnd.ms-powerpoint":                                             "ppt",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",

	// Archives.
	"application/zip":              "zip",
	"application/x-zip-compressed": "zip",
	"application/x-rar-compressed": "rar",
	"application/x-7z-compressed":  "7z",
	"application/x-tar":            "tar",
	"application/x-gzip":           "gz",
	"application/x-bzip2":          "bz2",
	"application/x-xz":             "xz",

	// Text and data.
	"text/csv":                    "csv",
	"application/csv":             "csv",
	"text/comma-separated-values": "csv",
	"application/json":            "json",
	"text/json":                   "json",
	"text/plain":                  "txt",
}

// getExt resolves the extension an accepted content type is stored under,
// preferring the caller's own table over the built-in one.
func getExt(mimeType string, override map[string]string) (string, error) {

	if ext, ok := override[mimeType]; ok {
		return ext, nil
	}

	if ext, ok := extensions[mimeType]; ok {
		return ext, nil
	}

	return "", ErrInvalidFileType.Var(mimeType)
}
