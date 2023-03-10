package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("helloworld"),
		fuse.Subtype("hellofs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}
}

// FS implements the hello world file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello.mp4" {
		return File{}, nil
	}
	return nil, syscall.ENOENT
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: "hello.mp4", Type: fuse.DT_File},
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dirDirs, nil
}

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0o444
	// a.Size = uint64(len(greeting))
	a.Size = uint64(size)
	return nil
}

var (
	size      int64 = 0
	largefile *os.File
)

func init() {
	var err error
	largefile, err = os.Open("/home/rok/Downloads/largefile.mp4")
	if err != nil {
		log.Fatal(err)
	}

	stat, err := largefile.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size = stat.Size()
}

func (File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	b := make([]byte, req.Size)
	log.Println(req.Offset, req.Size)
	_, err := largefile.ReadAt(b, req.Offset)
	if err != nil && err != io.EOF {
		log.Println(err)
		return err
	}

	resp.Data = b

	return nil
}
