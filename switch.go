package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Switch is a mode for switching files.
type Switch struct {
	editor    *Editor
	query     string
	paths     Paths
	selection Paths
	area      Area
	position  Position
}

// Show updates mode when switched to.
func (mode *Switch) Show() error {
	var err error
	mode.paths, err = mode.read()
	if err != nil {
		return fmt.Errorf("error showing switch mode: %w", err)
	}
	sort.Sort(mode.paths)
	mode.filter()
	return nil
}

// Hide updates mode when switched from.
func (mode *Switch) Hide() error {
	return nil
}

// Key handles input events.
func (mode *Switch) Key(key Key) error {
	var err error
	switch key {
	case KeyTab:
		mode.editor.SwitchMode(mode.editor.Command)
	case KeyUp:
		mode.moveUp()
	case KeyDown:
		mode.moveDown()
	case KeyLeft:
		mode.moveLeft()
	case KeyRight:
		mode.moveRight()
	case KeyBackspace:
		mode.delete()
		mode.filter()
	case KeyEnter:
		err = mode.open()
		mode.editor.SwitchMode(mode.editor.Command)
	}
	if err != nil {
		return fmt.Errorf("error handling key %v: %w", key, err)
	}
	return nil
}

// Rune handles rune input.
func (mode *Switch) Rune(rune rune) error {
	mode.append(rune)
	mode.filter()
	return nil
}

// Render renders mode.
func (mode *Switch) Render(view *View) error {
	mode.area = mode.area.Resize(view.Size).Shift(mode.position)
	selection := len(mode.selection) > 0
	for line := mode.area.T; line < mode.area.B; line++ {
		rline := line - mode.area.T
		for col := mode.area.L; col < mode.area.R; col++ {
			rcol := col - mode.area.L
			if line < len(mode.selection) {
				selected := []rune(mode.selection[line])
				if col < len(selected) {
					view.Content[rline][rcol] = selected[col]
				}
			}
			if selection && line == mode.position.L {
				view.Selection[rline][rcol] = true
			}
		}
	}
	status, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}
	view.Color = ColorBlue
	view.Position = mode.position
	view.Status = status
	view.Prompt = string(mode.query)
	view.Cursor = CursorPrompt
	return nil
}

func (mode *Switch) filter() {
	query := mode.query
	mode.selection = make([]string, 0, len(mode.paths))
	for _, path := range mode.paths {
		if mode.match(path, query) {
			mode.selection = append(mode.selection, path)
		}
	}
	mode.position = Position{}
	return
}

func (mode *Switch) open() error {
	pos := mode.position
	path := mode.query
	if pos.L < len(mode.selection) {
		path = mode.selection[pos.L]
	}
	err := mode.editor.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", path, err)
	}
	return nil
}

func (mode *Switch) append(rune rune) {
	mode.query = mode.query + string(rune)
}

func (mode *Switch) delete() {
	length := len(mode.query)
	if length != 0 {
		mode.query = mode.query[:length-1]
	}
}

func (mode *Switch) moveUp() {
	if mode.position.L > 0 {
		mode.position.L--
	}
}

func (mode *Switch) moveDown() {
	if mode.position.L+1 < len(mode.selection) {
		mode.position.L++
	}
}

func (mode *Switch) moveLeft() {
	if mode.area.L > 0 {
		mode.position.C = mode.area.L - 1
	} else {
		mode.position.C = 0
	}
}

func (mode *Switch) moveRight() {
	mode.position.C = mode.area.R + 1
}

func (mode *Switch) read() (paths []string, err error) {
	work, err := os.Getwd()
	if err != nil {
		return paths, fmt.Errorf("error reading working directory: %w", err)
	}
	walker := func(path string, info os.FileInfo, err error) error {
		relpath, err := filepath.Rel(work, path)
		if err != nil {
			return err
		}
		if info == nil {
			return nil
		}
		if info.Mode().IsRegular() {
			paths = append(paths, relpath)
		}
		return nil
	}
	err = filepath.Walk(work, walker)
	if err != nil {
		return paths, fmt.Errorf("error walking directory %s: %w", work, err)
	}
	return paths, nil
}

func (mode *Switch) match(path, query string) bool {
	if len(query) == 0 {
		return true
	}
	j := 0
	runes := []rune(query)
	for _, p := range path {
		q := runes[j]
		if p == q {
			j++
		}
		if j == len(query) {
			return true
		}
	}
	return false
}
