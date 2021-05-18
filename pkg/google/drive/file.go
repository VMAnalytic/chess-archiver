package drive

import (
	"io"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/drive/v3"
)

type File struct {
	ID          string
	Name        string
	Description string
	Tags        map[string]string
	Media       io.Reader

	UploadedAt *time.Time
	ModifiedAt *time.Time
	Author     *string
	EditedBy   *string
}

func newFileFromOrigin(file *drive.File) (*File, error) {
	if nil == file {
		return nil, nil
	}

	f := &File{
		ID:   file.Id,
		Name: file.Name,
	}

	if file.CreatedTime != "" {
		t, err := time.Parse(time.RFC3339, file.CreatedTime)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		f.UploadedAt = &t
	}

	if file.ModifiedTime != "" {
		t, err := time.Parse(time.RFC3339, file.ModifiedTime)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		f.ModifiedAt = &t
	}

	if file.SharingUser != nil {
		f.Author = &file.SharingUser.DisplayName
	}

	if file.LastModifyingUser != nil {
		f.EditedBy = &file.LastModifyingUser.DisplayName
	}

	return f, nil
}

func (f File) CreatedAt() *time.Time {
	return f.UploadedAt
}

func (f *File) AddTag(key, value string) {
	f.Tags[key] = value
}
