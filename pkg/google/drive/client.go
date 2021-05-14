package drive

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/pkg/errors"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	MimeTypeFolder  = "application/vnd.google-apps.folder"
	OrderDirection  = "createdTime desc" //sort by uploading time
	DefaultPageSize = 200                //should be less than 1000
)

type ErrGDrive struct {
	reason string
}

func NewErrGDrive(err error) *ErrGDrive {
	return &ErrGDrive{reason: errors.Wrap(err, "error during the request to GDrive API:").Error()}
}

func (e ErrGDrive) Error() string {
	return e.reason
}

type GDriveClient interface {
	//Get info about the file or folder by ID
	Get(ctx context.Context, ID string) (*File, error)

	//Files returns list of the files by IDs
	Files(ctx context.Context, IDs []string) ([]*File, error)

	//FilesFromFolder will return the list of files in folder.
	//recursively option provide possibility to get all files from sub folders
	FilesFromFolder(ctx context.Context, folderName string, recursively bool) ([]*File, error)

	//Latest will return the file in the folder which was last uploaded
	Latest(ctx context.Context, folderID string) (*File, error)

	//Folders return the list of the folders
	Folders(ctx context.Context) ([]*File, error)

	//Create create file
	Create(ctx context.Context, folder string, file *File) (string, error)
}

type HTTPClient struct {
	ds          *drive.Service
	rateLimiter *rate.Limiter
}

func NewHTTPtClient(ctx context.Context) (*HTTPClient, error) {
	client, err := drive.NewService(
		ctx,
		option.WithScopes(drive.DriveScope),
	)
	if err != nil {
		return nil, NewErrGDrive(err)
	}

	rl := rate.NewLimiter(rate.Every(1*time.Second), 10)

	return &HTTPClient{ds: client, rateLimiter: rl}, nil
}

func (m HTTPClient) Get(ctx context.Context, ID string) (*File, error) {
	f, err := m.ds.Files.
		Get(ID).
		Context(ctx).
		Fields("id, name, createdTime, modifiedTime, sharingUser, lastModifyingUser").
		Do()

	if err != nil {
		return nil, NewErrGDrive(err)
	}

	return newFileFromOrigin(f)
}

func (m *HTTPClient) Files(ctx context.Context, IDs []string) ([]*File, error) {
	var (
		next      = true
		pageToken string
		list      []*File
	)

	for next {
		r, err := m.ds.Files.
			List().
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).
			Corpora("allDrives").
			Fields("nextPageToken, files(id, name, createdTime, modifiedTime, sharingUser, lastModifyingUser)").
			PageSize(DefaultPageSize).
			OrderBy(OrderDirection).
			PageToken(pageToken).
			Q(fmt.Sprintf("mimeType!='%s'", MimeTypeFolder)).
			Do()
		if err != nil {
			return nil, NewErrGDrive(err)
		}

		originList := r.Files

		for _, f := range originList {
			file, err := newFileFromOrigin(f)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if !stringInSlice(f.Id, IDs) {
				continue
			}

			list = append(list, file)
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			next = false
		}
	}

	return list, nil
}

func (m *HTTPClient) Path(ctx context.Context, ID string) (string, error) {
	path, err := m.path(ctx, ID, "")

	return path, err
}

func (m *HTTPClient) path(ctx context.Context, ID string, path string) (string, error) {
	f, err := m.ds.Files.
		Get(ID).
		SupportsAllDrives(true).
		Context(ctx).
		Fields("id, name, parents").
		Do()

	if err != nil {
		return "", NewErrGDrive(err)
	}

	if len(f.Parents) == 0 {
		return path, nil
	}

	var r string

	p := f.Parents[0]
	p2, err := m.ds.Files.
		Get(ID).
		SupportsAllDrives(true).
		Context(ctx).
		Fields("id, name, parents").
		Do()

	if err != nil {
		return "", NewErrGDrive(err)
	}

	r, err = m.path(ctx, p, r)
	if err != nil {
		return "", NewErrGDrive(err)
	}

	r = r + "/" + p2.Name

	return r, err
}

func (m HTTPClient) FilesFromFolder(ctx context.Context, folderName string, recursively bool) ([]*File, error) {
	// Get all files from directory without subdirectories
	if !recursively {
		var (
			next      = true
			pageToken string
			list      []*File
		)

		for next {
			r, err := m.ds.Files.
				List().
				SupportsAllDrives(true).
				IncludeItemsFromAllDrives(true).
				Context(ctx).
				Fields("nextPageToken, files(id, name, createdTime, modifiedTime, sharingUser, lastModifyingUser)").
				PageSize(DefaultPageSize).
				OrderBy(OrderDirection).
				PageToken(pageToken).
				Q(fmt.Sprintf("mimeType!='%s' and '%s' in parents", MimeTypeFolder, folderName)).
				Do()
			if err != nil {
				return nil, NewErrGDrive(err)
			}

			originList := r.Files

			for _, f := range originList {
				file, err := newFileFromOrigin(f)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				list = append(list, file)
			}

			pageToken = r.NextPageToken
			if pageToken == "" {
				next = false
			}
		}

		return list, nil
	}

	// Get all files from directory with subdirectories
	if recursively {
		folders, err := m.SubFolders(ctx, folderName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var (
			sb        strings.Builder
			list      []*File
			pageToken string
			next      = true
		)

		list, err = m.FilesFromFolder(ctx, folderName, false)
		if err != nil {
			return nil, NewErrGDrive(err)
		}

		if len(folders) == 0 {
			return list, nil
		}

		sb.WriteString(fmt.Sprintf("'%s' in parents", folders[0].ID))

		for _, f := range folders[1:] {
			sb.WriteString(fmt.Sprintf(" or '%s' in parents", f.ID))
		}

		s := sb.String()

		for next {
			r, err := m.ds.Files.
				List().
				SupportsAllDrives(true).
				IncludeItemsFromAllDrives(true).
				Context(ctx).
				Fields("nextPageToken, files(*)").
				PageSize(DefaultPageSize).
				OrderBy(OrderDirection).
				PageToken(pageToken).
				Q(fmt.Sprintf("mimeType!='%s' and %s", MimeTypeFolder, s)).
				Do()
			if err != nil {
				return nil, NewErrGDrive(err)
			}

			originList := r.Files

			for _, f := range originList {
				file, err := newFileFromOrigin(f)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				list = append(list, file)
			}

			pageToken = r.NextPageToken
			if pageToken == "" {
				next = false
			}
		}

		return list, nil
	}

	return nil, nil
}

func (m HTTPClient) SubFolders(ctx context.Context, dirID string) ([]*File, error) {
	var (
		next      = true
		pageToken string
		sb        strings.Builder
		list      []*File
	)

	sb.WriteString(fmt.Sprintf("mimeType='%s'", MimeTypeFolder))

	if dirID != "" {
		sb.WriteString(fmt.Sprintf(" and '%s' in parents", dirID))
	}

	for next {
		r, err := m.ds.Files.
			List().
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).
			Fields("nextPageToken, files(id, name, createdTime, modifiedTime)").
			PageSize(DefaultPageSize).
			PageToken(pageToken).
			Q(sb.String()).
			Do()
		if err != nil {
			return nil, NewErrGDrive(err)
		}

		originList := r.Files

		for _, f := range originList {
			file, err := newFileFromOrigin(f)
			if err != nil {
				return nil, NewErrGDrive(err)
			}

			list = append(list, file)
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			next = false
		}
	}

	return list, nil
}

func (m HTTPClient) All(ctx context.Context) ([]*File, error) {
	folders, err := m.Folders(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var (
		sb   strings.Builder
		list []*File
	)

	if len(folders) == 0 {
		return list, nil
	}

	sb.WriteString(fmt.Sprintf("'%s' in parents", folders[0].ID))

	for _, f := range folders[1:] {
		sb.WriteString(fmt.Sprintf(" or '%s' in parents", f.ID))
	}

	s := sb.String()
	r, err := m.ds.Files.
		List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Context(ctx).
		Fields("*").
		PageSize(DefaultPageSize).
		OrderBy(OrderDirection).
		//PageToken(pageToken).
		Q(s).
		Q(fmt.Sprintf("mimeType!='%s'", MimeTypeFolder)).
		Do()

	if err != nil {
		return nil, NewErrGDrive(err)
	}

	originList := r.Files

	for _, f := range originList {
		file, err := newFileFromOrigin(f)
		if err != nil {
			return nil, err
		}

		list = append(list, file)
	}

	return list, nil
}

func (m HTTPClient) Folders(ctx context.Context) ([]*File, error) {
	var (
		next      = true
		pageToken string
		list      []*File
	)

	for next {
		r, err := m.ds.Files.
			List().
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).
			Fields("nextPageToken, files(id, name, createdTime, modifiedTime)").
			PageSize(DefaultPageSize).
			PageToken(pageToken).
			Q(fmt.Sprintf("mimeType='%s'", MimeTypeFolder)).
			Do()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		originList := r.Files

		for _, f := range originList {
			file, err := newFileFromOrigin(f)
			if err != nil {
				return nil, NewErrGDrive(err)
			}

			list = append(list, file)
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			next = false
		}
	}

	return list, nil
}

func (m HTTPClient) Create(ctx context.Context, folder string, file *File) (string, error) {
	err := m.rateLimiter.Wait(ctx)
	if err != nil {
		return "", errors.WithStack(err)
	}

	f := &drive.File{Name: file.Name, Properties: map[string]string{}}
	f.Parents = []string{folder}
	f.Properties = file.Tags
	f.Description = file.Description

	r, err := m.ds.Files.
		Create(f).
		Media(file.Media).
		Context(ctx).
		Do()

	if err != nil {
		return "", errors.WithStack(err)
	}

	return r.Id, nil
}

func (m HTTPClient) Latest(ctx context.Context, folderID string) (*File, error) {
	files, err := m.FilesFromFolder(ctx, folderID, false)
	if err != nil {
		return nil, NewErrGDrive(err)
	}

	if len(files) == 0 {
		return nil, nil
	}

	return files[0], err
}

func stringInSlice(ID string, IDList []string) bool {
	for _, ident := range IDList {
		if ident == ID {
			return true
		}
	}

	return false
}
