package tmpl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func templateFile(src, dst string, data interface{}) (err error) {
	_, err = os.Stat(src)
	if err != nil {
		return
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	tmpl, err := template.New(filepath.Base(src)).ParseFiles(src)
	if err != nil {
		fmt.Println("parse")
		return
	}

	err = tmpl.ExecuteTemplate(dstFile, filepath.Base(src), data)
	if err != nil {
		fmt.Println("Execute template error")
		return
	}

	return
}

// ApplyDir applies template data to a whole directory and copies the final output to destination dir
func ApplyDir(srcDir string, destDir string, data interface{}) error {
	e := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		destPath := destDir + strings.TrimPrefix(path, srcDir)

		fi, err := os.Stat(path)
		if fi.Mode().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			return nil
		} else if !fi.Mode().IsRegular() {
			// Not a file not a dir skip
			return nil
		}

		fmt.Printf("%s\n", destPath)
		err = templateFile(path, destPath, data)
		return err
	})

	if e != nil {
		os.RemoveAll(destDir)
		return e
	}
	return nil
}
