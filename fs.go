package main

import (
	"context"
	"database/sql"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS represents the top level filing system
type FS struct {
	db *sql.DB
}

// Check interface satistfied
var _ fs.FS = (*FS)(nil)

// NewFS makes a new FS
func NewFS(db *sql.DB) *FS {
	fsys := &FS{db: db}
	return fsys
}

// Root returns the root node
func (f *FS) Root() (node fs.Node, err error) {
	return &Dir{fs: f, path: "/"}, nil
}

// Check interface satsified
var _ fs.FSStatfser = (*FS)(nil)

// Statfs is called to obtain file system metadata.
// It should write that data to resp.
func (f *FS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) (err error) {
	const blockSize = 4096
	const fsBlocks = (1 << 50) / blockSize
	resp.Blocks = fsBlocks  // Total data blocks in file system.
	resp.Bfree = fsBlocks   // Free blocks in file system.
	resp.Bavail = fsBlocks  // Free blocks in file system if you're not root.
	resp.Files = 1E9        // Total files in file system.
	resp.Ffree = 1E9        // Free files in file system.
	resp.Bsize = blockSize  // Block size
	resp.Namelen = 255      // Maximum file name length?
	resp.Frsize = blockSize // Fragment size, smallest addressable data size in the file system.
	return nil
}