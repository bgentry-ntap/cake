package vsphere

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// CreateVMFolder creates all folders in a path like "one/two/three"
func (s *Session) CreateVMFolder(folderPath string) (map[string]*object.Folder, error) {
	folders := strings.Split(folderPath, "/")
	desiredFolders := make(map[string]*object.Folder)

	base, err := s.createVMFolderRootLevel(folders[0])
	if err != nil {
		return nil, err
	}
	desiredFolders[folders[0]] = base

	for f := 1; f < len(folders); f++ {
		nested, err := s.createVMFolderNestedLevel(desiredFolders[folders[f-1]], folders[f])
		if err != nil {
			return nil, err
		}
		desiredFolders[folders[f]] = nested
	}

	return desiredFolders, nil
}

// createVMFolderRootLevel creates a VM folder at the root level
func (s *Session) createVMFolderRootLevel(folderName string) (*object.Folder, error) {
	d := time.Now().Add(2 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	finder := find.NewFinder(s.Conn.Client, true)
	finder.SetDatacenter(s.Datacenter)
	iPath := s.Datacenter.InventoryPath + "/vm/" + folderName
	desiredFolder, err := finder.Folder(ctx, iPath)
	if err != nil {
		rootFolder := s.Datacenter.InventoryPath + "/vm"
		folder, err := finder.Folder(ctx, rootFolder)
		if err != nil {
			return nil, fmt.Errorf("unable to find folder, %v", err)
		}
		desiredFolder, err = folder.CreateFolder(ctx, folderName)
		if err != nil {
			return nil, fmt.Errorf("unable to create folder, %v", err)
		}
		if desiredFolder.InventoryPath == "" {
			desiredFolder.SetInventoryPath(iPath)
		}
	}

	return desiredFolder, nil
}

// createVMFolderNestedLevel creates a VM folder inside of a root level folder
func (s *Session) createVMFolderNestedLevel(rootFolder *object.Folder, folderName string) (*object.Folder, error) {
	d := time.Now().Add(2 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	finder := find.NewFinder(s.Conn.Client, true)
	finder.SetDatacenter(s.Datacenter)
	desiredFolder := new(object.Folder)
	n := fmt.Sprintf("%s/%s", rootFolder.InventoryPath, folderName)
	desiredFolder, err := finder.Folder(ctx, n)
	if err != nil {
		desiredFolder, err = rootFolder.CreateFolder(ctx, folderName)
		if err != nil {
			return nil, fmt.Errorf("unable to create folder, %v", err)
		}
	}
	if desiredFolder.InventoryPath == "" && rootFolder.InventoryPath != "" {
		desiredFolder.SetInventoryPath(rootFolder.InventoryPath + "/" + folderName)
	}

	return desiredFolder, nil
}

// DeleteVMFolder removes a folder from vcenter
func (s *Session) DeleteVMFolder(folder *object.Folder) (*object.Task, error) {
	d := time.Now().Add(2 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	var task *object.Task
	finder := find.NewFinder(s.Conn.Client, true)
	finder.SetDatacenter(s.Datacenter)
	found, err := finder.Folder(ctx, folder.InventoryPath)
	if err == nil {
		task, err = found.Destroy(ctx)
		if err != nil {
			return nil, err
		}
	}

	return task, nil
}
