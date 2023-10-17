package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type File struct{
	path string
	fs *FS
}

// Check interface satisfied
var _ fs.Node = (*File)(nil)

// TODO: cache for `sqlToCsv`

var contentCache = make(map[string]string)

func fileContentToCsv(f *File) (string, error) {
	if cached, found := contentCache[f.path]; found {
		fmt.Println("Use content from cache:", f.path)
		return cached, nil
	}

	sb := strings.Builder{}

	// f.path = "/dbname/tblname/by/id/2

	infos := strings.Split(f.path[1:], "/")

	f.fs.db.Exec("USE " + infos[0])

	var sql string

	if len(infos) == 3 && infos[2] == "all" {
		sql = fmt.Sprintf("SELECT * FROM `%s`", infos[1])
	} else {
		sql = fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` = '%s'", infos[1], infos[3], infos[4])
	}

	rows, err := f.fs.db.Query(sql)
	if err != nil {
		log.Printf("ERR 1 %s\n", err.Error())
		return "", err
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("ERR 2 %s\n", err.Error())
		return "", err
	}

	// write header
	sb.WriteString(strings.Join(cols, ","))
	sb.WriteRune('\n')

	// Create a slice of interface{} to store the values
	values := make([]interface{}, len(cols))
	for i := range values {
		var v interface{}
		values[i] = &v
	}

	i := 0
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			log.Printf("ERR 3 %s\n", err.Error())
			return "", err
		}

		// Iterate through the retrieved values
		for i := range cols {
			val := *values[i].(*interface{})
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(formatValue(val))
		}
		sb.WriteRune('\n')

		i++
	}

	contentCache[f.path] = sb.String()

	return contentCache[f.path], nil
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("File.Attr(%#v)\n", f.path)

	if strings.Contains(f.path, ".git") {
		return syscall.ENOENT
	}

	content, err := fileContentToCsv(f)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		return syscall.ENOENT
	}

	a.Mode = 0o444
	a.Size = uint64(len(content))
	a.Atime = time.Now()
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Printf("File.ReadAll(%#v)\n", f.path)
	
	content, err := fileContentToCsv(f)
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

// Check interface satisfied
var _ fs.NodeFsyncer = (*File)(nil)

// Fsync the file
//
// Note that we don't do anything except return OK
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) (err error) {
	return nil
}