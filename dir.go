package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var (
	ErrorDbNotFound = errors.New("db not found")
)

type Dir struct{
	fs *FS
	path string
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	// a.Inode = 0 // todo
	a.Mode = os.ModeDir | 0o555
	return nil
}

// Check interface satisfied
var _ fs.NodeRequestLookuper = (*Dir)(nil)

// Lookup looks up a specific entry in the receiver.
//
// Lookup should return a Node corresponding to the entry.  If the
// name does not exist in the directory, Lookup should return ENOENT.
//
// Lookup need not to handle the names "." and "..".
func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (node fs.Node, err error) {
	var fullpath string

	if d.path == "/" {
		fullpath = "/" + req.Name
	} else {
		fullpath = d.path + "/" + req.Name
	}

	infos := strings.Split(fullpath[1:], "/")

	log.Printf("Dir.Lookup(%s) %+v\n", fullpath, infos)

	if strings.HasPrefix(req.Name, ".Trash") {
		return nil, syscall.ENOENT
	}

	if len(infos) == 5 && infos[2] == "by" {
		return &File{fs: d.fs, path: fullpath}, nil
	}

	if len(infos) == 3 && infos[2] == "all" {
		return &File{fs: d.fs, path: fullpath}, nil
	}

	return &Dir{fs: d.fs, path: fullpath}, nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Printf("Dir.ReadDirAll(%s)\n", d.path)

	dirList := make([]fuse.Dirent, 0)

	infos := strings.Split(d.path[1:], "/")

	fmt.Printf("Asking for: %#v\n", infos)

	if strings.HasPrefix(infos[0], ".Trash-") {
		return dirList, nil
	}

	if len(infos) == 0 {
		return dirList, nil
	}

	databases := make([]string, 0)
	foundCurrentDb := false

	if len(infos) > 0 {
		rows, err := d.fs.db.Query("SHOW DATABASES")

		if err != nil {
			log.Printf("ERR 1 %s\n", err.Error())
			return dirList, err
		}

		defer rows.Close()

		for rows.Next() {
			var dbname string
			if err := rows.Scan(&dbname); err != nil {
				log.Printf("ERR 2 %s\n", err.Error())
				return dirList, err
			}
			databases = append(databases, dbname)

			if dbname == infos[0] {
				foundCurrentDb = true
			}
		}
	}

	if d.path == "/" {
		for _, db := range databases {
			dirList = append(dirList, fuse.Dirent{Name: db, Type: fuse.DT_Dir})
		}

		return dirList, nil
	}

	if !foundCurrentDb {
		return dirList, ErrorDbNotFound
	}

	d.fs.db.Exec("USE " + infos[0])

	if len(infos) == 1 {
		// we have only the name of a database
		// d.path = "/dbname

		rows, err := d.fs.db.Query("SHOW TABLES")
		if err != nil {
			log.Printf("ERR 5 %s\n", err.Error())
			return dirList, err
		}

		defer rows.Close()

		for rows.Next() {
			var tblname string
			if err := rows.Scan(&tblname); err != nil {
				log.Printf("ERR 6 %s\n", err.Error())
				return dirList, err
			}
			dirList = append(dirList, fuse.Dirent{Name: tblname, Type: fuse.DT_Dir})
		}

		return dirList, nil
	}

	// TODO: check if infos[0] is a valid DB name

	if len(infos) == 2 {
		dirList = append(dirList, fuse.Dirent{Name: "by", Type: fuse.DT_Dir})
		dirList = append(dirList, fuse.Dirent{Name: "all", Type: fuse.DT_File})

		return dirList, nil
	}

	if len(infos) == 3 && infos[2] == "by" {
		// we have the name of a database, a table
		// d.path = "/dbname/tblname/by

		rows, err := d.fs.db.Query("SELECT * FROM " + infos[1] + " LIMIT 1")
		if err != nil {
			log.Printf("ERR 7 %s\n", err.Error())
			return dirList, err
		}

		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			log.Printf("ERR 4 %s\n", err.Error())
			return dirList, err
		}

		for _, colName := range cols {
			dirList = append(dirList, fuse.Dirent{Name: colName, Type: fuse.DT_Dir})
		}
		return dirList, nil
	}

	if len(infos) == 4 && infos[2] == "by" {
		// we have the name of a database, a table and a column
		// d.path = "/dbname/tblname/by/id

		sql := fmt.Sprintf("SELECT DISTINCT `%s` as `val` FROM `%s` ORDER BY `%s`", infos[3], infos[1], infos[3])
			rows, err := d.fs.db.Query(sql)
		if err != nil {
			log.Printf("ERR 7 %s\n", err.Error())
			return dirList, err
		}

		defer rows.Close()

		for rows.Next() {
			var colValue string
			if err := rows.Scan(&colValue); err != nil {
				log.Printf("ERR 6 %s\n", err.Error())
				return dirList, err
			}
			dirList = append(dirList, fuse.Dirent{Name: colValue, Type: fuse.DT_Dir})
		}

		return dirList, nil
	}

	if len(infos) == 5 && infos[2] == "by" {
		// we have the name of a database, a table, a column and a value for that column
		// d.path = "/dbname/tblname/by/id/2

		sql := fmt.Sprintf("SELECT * FROM `%s` WHERE `%s` = %s", infos[1], infos[3], infos[4])
		rows, err := d.fs.db.Query(sql)
		if err != nil {
			log.Printf("ERR 7 %s\n", err.Error())
			return dirList, err
		}

		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			log.Printf("ERR 4 %s\n", err.Error())
			return dirList, err
		}

		// Create a slice of interface{} to store the values
		values := make([]interface{}, len(cols))
		for i := range values {
			var v interface{}
			values[i] = &v
		}

		i := 0
		for rows.Next() {
			if err := rows.Scan(values...); err != nil {
				log.Printf("ERR 8 %s\n", err.Error())
				return dirList, err
			}

			// Iterate through the retrieved values
			for i, colName := range cols {
				val := *values[i].(*interface{})
				fmt.Printf("%s: %v\n", colName, formatValue(val))
				dirList = append(dirList, fuse.Dirent{Name: colName, Type: fuse.DT_File})
			}
			fmt.Println("---------------")

			i++
		}

		return dirList, nil
	}

	return dirList, nil
}

// Format values as a readable string
func formatValue(val interface{}) string {
	switch t := val.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}
