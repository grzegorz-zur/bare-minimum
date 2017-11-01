package bm

import (
	"bufio"
	tb "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"log"
	"os"
)

type File struct {
	Path     string
	Data     [][]rune
	Window   Bounds
	Position Position
}

func OpenFile(path string) (file *File, err error) {
	file = &File{
		Path: path,
	}
	f, err := os.Open(path)
	if err != nil {
		err = errors.Wrapf(err, "cannot open file: %s", path)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			err = errors.Wrapf(err, "cannot close file: %s", path)
			log.Print(err)
		}
	}()
	s := bufio.NewScanner(f)
	for s.Scan() {
		err = s.Err()
		if err != nil {
			err = errors.Wrapf(err, "cannot read file: %s", path)
			return
		}
		line := s.Text()
		runes := []rune(line)
		file.Data = append(file.Data, runes)
	}
	return
}

func (file *File) Write() (err error) {
	f, err := os.Create(file.Path)
	if err != nil {
		err = errors.Wrapf(err, "cannot write file: %s", file.Path)
		return
	}
	for i, runes := range file.Data {
		line := string(runes)
		if i+1 < len(file.Data) {
			line += "\n"
		}
		bytes := []byte(line)
		f.Write(bytes)
	}
	return
}

func (file *File) Resize(size Size) {
	file.Window.Bottom = file.Window.Top + size.Lines
	file.Window.Right = file.Window.Left + size.Cols
	if file.Position.Line > file.Window.Bottom {
		file.Position.Line = file.Window.Bottom
	}
	if file.Position.Col > file.Window.Right {
		file.Position.Col = file.Window.Right
	}
	return
}

func (file *File) Display(position Position) (cursor Position) {
	for line := file.Window.Top; line <= file.Window.Bottom; line++ {
		if line >= len(file.Data) {
			break
		}
		runes := file.Data[line]
		absLine := position.Line + line
		for col := file.Window.Left; col <= file.Window.Right; col++ {
			if col >= len(runes) {
				break
			}
			symbol := runes[col]
			absCol := position.Col + col
			tb.SetCell(absCol, absLine, symbol, tb.ColorDefault, tb.ColorDefault)
		}
	}
	cursor = file.Position
	return
}

func (file *File) MoveLeft() {
	p := &file.Position
	if p.Col > 0 {
		p.Col--
	}
}

func (file *File) MoveRight() {
	p := &file.Position
	p.Col++
}

func (file *File) MoveUp() {
	p := &file.Position
	if p.Line > 0 {
		p.Line--
	}
}

func (file *File) MoveDown() {
	p := &file.Position
	p.Line++
}

func (file *File) Insert(r rune) {
	if file.empty() {
		file.extend()
	} else {
		file.shiftRight()
	}
	file.extend()
	p := &file.Position
	file.Data[p.Line][p.Col] = r
	p.Col += 1
}

func (file *File) Delete() {
	switch {
	case file.empty():
		return
	case file.emptyLine():
		file.DeleteLine()
	case file.emptyChar():
		return
	default:
		file.DeleteChar()
	}
}

func (file *File) DeleteChar() {
	p := &file.Position
	line := &file.Data[p.Line]
	rest := (*line)[p.Col+1:]
	*line = append((*line)[:p.Col], rest...)
}

func (file *File) DeleteLine() {
	p := &file.Position
	data := &file.Data
	rest := (*data)[p.Line+1:]
	*data = append(*data, rest...)
}

func (file *File) empty() bool {
	p := &file.Position
	return p.Line >= len(file.Data) ||
		p.Col >= len(file.Data[p.Line])
}

func (file *File) emptyLine() bool {
	p := &file.Position
	line := &file.Data[p.Line]
	return len(*line) == 0
}

func (file *File) emptyChar() bool {
	p := &file.Position
	return p.Col >= len(file.Data[p.Line])
}

func (file *File) extend() {
	file.extendLine()
	file.extendCol()
}

func (file *File) extendLine() {
	p := &file.Position
	data := &file.Data
	for p.Line >= len(*data) {
		*data = append(*data, []rune{})
	}
}

func (file *File) extendCol() {
	p := &file.Position
	line := &file.Data[p.Line]
	for p.Col >= len(*line) {
		*line = append(*line, ' ')
	}
}

func (file *File) shiftRight() {
	p := &file.Position
	line := &file.Data[p.Line]
	rest := append([]rune{' '}, (*line)[p.Col:]...)
	*line = append((*line)[:p.Col], rest...)
}
