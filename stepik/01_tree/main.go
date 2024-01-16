package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

const (
	interlineSeparator = "├"
	finalSeparator     = "└"
)

func main() {
	out := os.Stdout

	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)

	if err != nil {
		panic(err.Error())
	}
}

// Представление размера файла в печатаемом виде
func sizeAsStr(f fs.DirEntry) string {
	inf, _ := f.Info()
	size := inf.Size()

	if size == 0 {
		return "(empty)"
	}

	return fmt.Sprintf("(%db)", inf.Size())
}

// Отфильтровывание папок в зависимости от условий
func filterEntries(dir []fs.DirEntry, printFiles bool) []fs.DirEntry {
	var entries []fs.DirEntry

	for _, f := range dir {
		if f.Name() == ".idea" || f.Name() == ".DS_Store" {
			continue
		}

		if !printFiles && !f.IsDir() {
			continue
		}

		entries = append(entries, f)
	}

	return entries
}

// Определение разделителей в псевдографике
func resolveGraphicSeparators(path string, isLast bool, lasts []bool) string {
	var separator string

	if isLast {
		separator = finalSeparator
	} else {
		separator = interlineSeparator
	}

	// пытаемся понять, на каком уровне вложенности мы сейчас находимся
	if strings.Contains(path, string(os.PathSeparator)) {
		c := strings.Count(path, string(os.PathSeparator))
		var separatorPrefix string

		for i := 0; i < c; i++ {
			if lasts[i] {
				separatorPrefix += "\t"
			} else {
				separatorPrefix += "│\t"
			}
		}

		separator = separatorPrefix + separator
	}

	return separator
}

// Вывод дерева
func printTree(out io.Writer, path string, printFiles bool, lasts []bool) error {
	lasts = append(lasts, false)
	dir, err := fs.ReadDir(os.DirFS(path), ".")

	if err != nil {
		return err
	}

	// фильтруем файлы, если надо
	entries := filterEntries(dir, printFiles)

	for idx, f := range entries {
		var isLast = idx == len(entries)-1
		var separator = resolveGraphicSeparators(path, isLast, lasts)

		lasts[(len(lasts) - 1)] = isLast

		if f.IsDir() {
			_, err := out.Write([]byte(fmt.Sprintf("%s───%s\n", separator, f.Name())))
			if err != nil {
				return err
			}

			// рекурсивный вызов
			err = printTree(out, path+string(os.PathSeparator)+f.Name(), printFiles, lasts)
			if err != nil {
				return err
			}
		} else if printFiles {
			_, err = out.Write([]byte(fmt.Sprintf("%s───%s %s\n", separator, f.Name(), sizeAsStr(f))))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Построение дерева
func dirTree(out io.Writer, path string, printFiles bool) error {
	err := printTree(out, path, printFiles, []bool{})

	if err != nil {
		return err
	}

	return nil
}
