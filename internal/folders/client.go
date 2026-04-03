package folders

import (
	"context"
	"errors"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Folder holds the fields we care about.
type Folder struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Parent      string `json:"parent"`
	State       string `json:"state"`
}

// Client defines operations for folders.
type Client interface {
	List(ctx context.Context, parent string) ([]*Folder, error)
	Get(ctx context.Context, folderID string) (*Folder, error)
	Create(ctx context.Context, parent, displayName string) (*Folder, error)
	Delete(ctx context.Context, folderID string) error
	Move(ctx context.Context, folderID, destParent string) (*Folder, error)
}

type gcpClient struct {
	rm *resourcemanager.FoldersClient
}

// NewClient creates a Client backed by the real GCP API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	rm, err := resourcemanager.NewFoldersClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create folders client: %w", err)
	}
	return &gcpClient{rm: rm}, nil
}

func (c *gcpClient) List(ctx context.Context, parent string) ([]*Folder, error) {
	it := c.rm.ListFolders(ctx, &resourcemanagerpb.ListFoldersRequest{Parent: parent})

	var folders []*Folder
	for {
		f, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list folders: %w", err)
		}
		folders = append(folders, fromProto(f))
	}
	return folders, nil
}

func (c *gcpClient) Get(ctx context.Context, folderID string) (*Folder, error) {
	f, err := c.rm.GetFolder(ctx, &resourcemanagerpb.GetFolderRequest{Name: "folders/" + folderID})
	if err != nil {
		return nil, fmt.Errorf("get folder %s: %w", folderID, err)
	}
	return fromProto(f), nil
}

func (c *gcpClient) Create(ctx context.Context, parent, displayName string) (*Folder, error) {
	op, err := c.rm.CreateFolder(ctx, &resourcemanagerpb.CreateFolderRequest{
		Folder: &resourcemanagerpb.Folder{
			Parent:      parent,
			DisplayName: displayName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}

	f, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for folder creation: %w", err)
	}
	return fromProto(f), nil
}

func (c *gcpClient) Delete(ctx context.Context, folderID string) error {
	op, err := c.rm.DeleteFolder(ctx, &resourcemanagerpb.DeleteFolderRequest{Name: "folders/" + folderID})
	if err != nil {
		return fmt.Errorf("delete folder %s: %w", folderID, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for folder deletion: %w", err)
	}
	return nil
}

func (c *gcpClient) Move(ctx context.Context, folderID, destParent string) (*Folder, error) {
	op, err := c.rm.MoveFolder(ctx, &resourcemanagerpb.MoveFolderRequest{
		Name:              "folders/" + folderID,
		DestinationParent: destParent,
	})
	if err != nil {
		return nil, fmt.Errorf("move folder %s: %w", folderID, err)
	}

	f, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for folder move: %w", err)
	}
	return fromProto(f), nil
}

func fromProto(f *resourcemanagerpb.Folder) *Folder {
	return &Folder{
		Name:        f.GetName(),
		DisplayName: f.GetDisplayName(),
		Parent:      f.GetParent(),
		State:       f.GetState().String(),
	}
}
