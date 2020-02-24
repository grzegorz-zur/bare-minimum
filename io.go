package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	perm os.FileMode = 0644
)

// Open opens a file.
func Open(path string) (file File, err error) {
	file = File{
		Path:    path,
		history: &History{},
	}
	_, err = file.Read(false)
	if err != nil {
		return file, fmt.Errorf("error opening file %s: %w", path, err)
	}
	return file, nil
}

// Read loads thie file.
func (f *File) Read(force bool) (read bool, err error) {
	if f == nil {
		return false, nil
	}
	stat, err := os.Stat(f.Path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("error reading file %s info: %w", f.Path, err)
	}
	exists := stat != nil
	changed := true
	if exists {
		changed = f.time != stat.ModTime()
	}
	if !changed && !force {
		return false, nil
	}
	flags := os.O_RDWR | os.O_CREATE
	fx, err := os.OpenFile(f.Path, flags, perm)
	if err != nil {
		return false, fmt.Errorf("error opening file %s: %w", f.Path, err)
	}
	defer fx.Close()
	buffer := &bytes.Buffer{}
	io.Copy(buffer, fx)
	f.content = string(buffer.Bytes())
	fx.Close()
	stat, err = os.Stat(f.Path)
	if err != nil {
		return true, fmt.Errorf("error reading file %s info: %w", f.Path, err)
	}
	f.changed = false
	f.time = stat.ModTime()
	f.Archive()
	return true, nil
}

// Write writes file contents.
func (f *File) Write() (wrote bool, err error) {
	if f == nil {
		return false, nil
	}
	stat, err := os.Stat(f.Path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("error reading file %s info: %w", f.Path, err)
	}
	exists := stat != nil
	changed := true
	if exists {
		changed = f.time != stat.ModTime()
	}
	if !f.changed && !changed {
		return false, nil
	}
	fx, err := os.Create(f.Path)
	defer fx.Close()
	if err != nil {
		return false, fmt.Errorf("error writing file %s: %w", f.Path, err)
	}
	bytes := []byte(f.content)
	fx.Write(bytes)
	fx.Close()
	stat, err = os.Stat(f.Path)
	if err != nil {
		return true, fmt.Errorf("error reading file %s info: %w", f.Path, err)
	}
	f.changed = false
	f.time = stat.ModTime()
	return true, nil
}
