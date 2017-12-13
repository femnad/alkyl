package main

import (
	"fmt"
	"os"
	"path"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"

	"./notifications"
)

const (
	TokenEnviromentVar = "ALKYL_GITHUB_TOKEN"
)

var Repo string

func main() {
	name := path.Base(os.Args[0])
	subType := fmt.Sprintf("%sfs", name)

	if len(os.Args) != 3 {
		fmt.Printf("usage: %s <mount-point> <repo>\n", name)
		os.Exit(1)
	}

	args := os.Args[1:]
	mountpoint := args[0]
	Repo = args[1]
	conn, err := fuse.Mount(mountpoint,
		fuse.FSName(name),
		fuse.Subtype(subType),
	)
	defer conn.Close()

	if err != nil {
		panic(err)
	}
	fs.Serve(conn, FS{})
}

type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{name: "/"}, nil
}

type Dir struct {
	name string
}

func (d Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir
	return nil
}

var dirEntries map[string][]notifications.Issue

func getAndMaybeStoreDirContents(name string) []notifications.Issue {
	issues, alreadyPresent := dirEntries[name]
	if !alreadyPresent {
		issues = notifications.GetIssues(Repo)
		if dirEntries == nil {
			dirEntries = make(map[string][]notifications.Issue)
		}
		dirEntries[name] = issues
	}
	return issues
}

func (d Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	issues := getAndMaybeStoreDirContents(d.name)
	for _, issue := range issues {
		if issue.Title == name {
			return File{issue: issue}, nil
		}
	}
	return nil, fuse.ENOENT
}

func GetDirEntries(issues []notifications.Issue) []fuse.Dirent {
	files := make([]fuse.Dirent, 0)
	for _, issue := range issues {
		dirEnt := fuse.Dirent{
			Inode: issue.Id,
			Name:  issue.Title,
			Type:  fuse.DT_File,
		}
		files = append(files, dirEnt)
	}
	return files
}

func (d Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	issues := getAndMaybeStoreDirContents(d.name)
	return GetDirEntries(issues), nil
}

type File struct {
	issue notifications.Issue
}

func (f File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.issue.Id
	a.Mode = 0444
	a.Size = uint64(len(f.issue.Body))
	return nil
}

func (f File) ReadAll(ctx context.Context) ([]byte, error) {
	fmt.Printf("readall with f: %s\n", f)
	return []byte(f.issue.Body), nil
}
