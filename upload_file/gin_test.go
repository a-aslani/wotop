package upload_file

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// A multipart body carrying one file part with an explicit content type, which
// is what the accept check and the extension lookup both read — not the
// filename's own extension.
func requestWithFile(t *testing.T, field, filename, contentType string, content []byte) *gin.Context {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/upload", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	return c
}

func emptyRequest(t *testing.T) *gin.Context {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.Close())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/upload", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	return c
}

func TestUpload_StoresAnAcceptedFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("a png, honestly")

	path, err := Upload(requestWithFile(t, "file", "photo.png", "image/png", content), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"image/png"},
	})

	require.NoError(t, err)
	require.True(t, strings.HasSuffix(path, ".png"), "stored path %q has no .png extension", path)

	stored, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, stored)
}

// An SVG is what a brand logo usually arrives as. The extension table had no
// entry for it, so it was refused as an invalid type however the handler
// declared its accepted types.
func TestUpload_StoresAnSvg(t *testing.T) {
	dir := t.TempDir()

	path, err := Upload(requestWithFile(t, "file", "mark.svg", "image/svg+xml", []byte(`<svg/>`)), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"image/svg+xml"},
	})

	require.NoError(t, err)
	require.True(t, strings.HasSuffix(path, ".svg"), "stored path %q has no .svg extension", path)
	require.FileExists(t, path)
}

// The built-in table can never be complete, so a caller can supply the mapping
// for a type it accepts rather than waiting for a release of this module.
func TestUpload_PrefersTheCallersExtensionTable(t *testing.T) {
	dir := t.TempDir()

	path, err := Upload(requestWithFile(t, "file", "book.epub", "application/epub+zip", []byte("epub")), Params{
		FieldName:  "file",
		Path:       dir,
		MaxSize:    1024,
		Accept:     []string{"application/epub+zip"},
		Extensions: map[string]string{"application/epub+zip": "epub"},
	})

	require.NoError(t, err)
	require.True(t, strings.HasSuffix(path, ".epub"), "stored path %q has no .epub extension", path)
}

// An upload directory is often shared: one volume, several services writing
// into it. A service that does not own the directory must still be able to write
// there, so the mode it was given is left alone.
//
// This is the failure the change came from — gin's SaveUploadedFile chmods the
// directory on every write since v1.11, which returns EPERM to a non-owner:
// "chmod uploads/x: operation not permitted".
func TestUpload_LeavesAnExistingDirectoryModeAlone(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes are not meaningful on Windows")
	}

	dir := filepath.Join(t.TempDir(), "shared")
	require.NoError(t, os.MkdirAll(dir, 0o777))
	require.NoError(t, os.Chmod(dir, 0o777))

	_, err := Upload(requestWithFile(t, "file", "photo.png", "image/png", []byte("png")), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"image/png"},
	})
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o777), info.Mode().Perm(), "the shared directory's mode was rewritten")
}

func TestUpload_CreatesAMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does", "not", "exist")

	path, err := Upload(requestWithFile(t, "file", "photo.png", "image/png", []byte("png")), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"image/png"},
	})

	require.NoError(t, err)
	require.FileExists(t, path)
}

func TestUpload_SkipsAnAbsentOptionalField(t *testing.T) {
	path, err := Upload(emptyRequest(t), Params{
		FieldName: "file",
		Path:      t.TempDir(),
		MaxSize:   1024,
		Accept:    []string{"image/png"},
	})

	require.NoError(t, err)
	require.Empty(t, path)
}

func TestUpload_RefusesAnAbsentRequiredField(t *testing.T) {
	_, err := Upload(emptyRequest(t), Params{
		FieldName:  "file",
		IsRequired: true,
		Path:       t.TempDir(),
		MaxSize:    1024,
		Accept:     []string{"image/png"},
	})

	require.ErrorIs(t, err, ErrMissingFile)
}

func TestUpload_RefusesAnUnacceptedType(t *testing.T) {
	dir := t.TempDir()

	_, err := Upload(requestWithFile(t, "file", "clip.gif", "image/gif", []byte("GIF89a")), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"image/png"},
	})

	require.Error(t, err)
	requireEmptyDir(t, dir)
}

// A type the handler accepts but the table cannot name is still refused — it
// just says so through the same error, which is why Params.Extensions exists.
func TestUpload_RefusesAnAcceptedTypeItCannotName(t *testing.T) {
	dir := t.TempDir()

	_, err := Upload(requestWithFile(t, "file", "data.bin", "application/octet-stream", []byte("bin")), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   1024,
		Accept:    []string{"application/octet-stream"},
	})

	require.Error(t, err)
	requireEmptyDir(t, dir)
}

func TestUpload_RefusesAnOversizedFile(t *testing.T) {
	dir := t.TempDir()

	_, err := Upload(requestWithFile(t, "file", "photo.png", "image/png", bytes.Repeat([]byte("x"), 11)), Params{
		FieldName: "file",
		Path:      dir,
		MaxSize:   10,
		Accept:    []string{"image/png"},
	})

	require.Error(t, err)
	requireEmptyDir(t, dir)
}

// SaveFileInDir wrote to params.Path joined to itself, with the extension's dot
// doubled, and then reported no path at all.
func TestFileUploader_SaveFileInDirReportsTheFileItWrote(t *testing.T) {
	dir := t.TempDir()
	content := []byte("a csv, honestly")

	uploader := NewUploader()
	err := uploader.Upload(requestWithFile(t, "file", "leads.CSV", "text/csv", content), Params{
		FieldName:     "file",
		Path:          dir,
		MaxSize:       1024,
		Accept:        []string{"text/csv"},
		SaveFileInDir: true,
	})
	require.NoError(t, err)

	path := uploader.Path()
	require.NotEmpty(t, path, "a file was written but no path was reported")
	require.True(t, strings.HasSuffix(path, ".csv"), "stored path %q has no .csv extension", path)
	require.NotContains(t, filepath.ToSlash(path), "..csv", "the extension's dot was doubled")

	stored, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, stored)

	// One file, directly in the directory asked for — not nested under a repeat
	// of that same path.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.False(t, entries[0].IsDir(), "the file was written into a nested copy of the upload path")

	require.Equal(t, int64(len(content)), uploader.Size())
	require.Equal(t, ".csv", uploader.Ext, "Ext keeps its leading dot, which callers match on")
}

func TestFileUploader_KeepsTheTempFileWhenAsked(t *testing.T) {
	tempDir := t.TempDir()
	pattern := "import-*.csv"

	uploader := NewUploader()
	err := uploader.Upload(requestWithFile(t, "file", "leads.csv", "text/csv", []byte("a,b")), Params{
		FieldName:   "file",
		Path:        t.TempDir(),
		MaxSize:     1024,
		Accept:      []string{"text/csv"},
		TempDir:     &tempDir,
		TempPattern: &pattern,
	})
	require.NoError(t, err)
	defer uploader.Close()

	require.NotNil(t, uploader.TempFile())

	// Rewound, so the caller reads from the start.
	stored, err := os.ReadFile(uploader.TempFile().Name())
	require.NoError(t, err)
	require.Equal(t, []byte("a,b"), stored)

	// Nothing was written to Path: SaveFileInDir was not asked for.
	require.Empty(t, uploader.Path())
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries, "a refused upload left files behind")
}
