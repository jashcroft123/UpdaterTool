package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func readLatestFromFolder(path string, remote bool, loc string) (string, int64, bool, error) {

	var latest, current int64
	var fileName string

	if !remote {
		path = path + "\\latest"
	}

	dir, err := os.Open(path)

	if err != nil {
		return "", 0, false, err
	}

	defer dir.Close()
	files, err := dir.ReadDir(-1)
	check(err)

	//Find latest version in remote folder
	for _, file := range files {
		fmt.Println(loc, ":\t", file.Name())
		_, after, found := strings.Cut(file.Name(), "@")
		before, _, found := strings.Cut(after, ".zip")
		_, _, forced := strings.Cut(file.Name(), "force")

		//Check if any item is forced, should only ever be one
		if forced {
			fmt.Println("Forced version found, updating to selected...")
			res := stringToVersion(before)
			current, _ = strconv.ParseInt(res, 10, 64)
			fileName = file.Name()
			return fileName, current, forced, err
		}

		//Allow non valid files to exist, these will be omitted
		if found {
			res := stringToVersion(before)
			current, _ = strconv.ParseInt(res, 10, 64)

			if current > latest {
				latest = current
				fileName = file.Name()
			}
		}
	}

	//Return path and latest version for each type i.e. Local, Remote
	return fileName, latest, false, err
}

// stringToVersion converts a string to a verion only string to be parsed to int
// Allow seperation of any type i.e. x_y_z, xyz, x-y etc, only numerics pass
func stringToVersion(str string) string {
	re := regexp.MustCompile(`[^0-9 ]+`)
	return re.ReplaceAllString(str, "")

}

// updateLatest grabs the zip file path and populates a new folder called
// "latest" with its contents
func updateLatest(src string, dst string) {
	archive, err := zip.OpenReader(src)
	check(err)
	defer archive.Close()

	for _, f := range archive.File {

		fp := filepath.Join(dst, f.Name)

		//Create directory
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fp, os.ModePerm)
			check(err)
			continue
		}

		//Create parent directory
		err := os.MkdirAll(filepath.Dir(fp), os.ModePerm)
		check(err)

		//Create destination file
		dstFile, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		check(err)

		srcFile, err := f.Open()
		check(err)

		_, err = io.Copy(dstFile, srcFile)
		check(err)

		dstFile.Close()
		srcFile.Close()
	}

}
