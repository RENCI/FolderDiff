package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/RENCI/GoUtils/Collections"
	"github.com/RENCI/GoUtils/FileSystem"
	"hash"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type FileAndHash struct {
	Path string
	Hash string
}

func main() {
	println("-------------Folder START-------------")
	defer timeTrack(time.Now(), "Total time")
	args := os.Args
	path1 := args[1]
	path2 := args[2]
	f1, total1, _ := getFiles(path1)
	f2, total2, _ := getFiles(path2)

	newfiles := Collections.NewList[FileAndHash]()
	updatedfiles := Collections.NewList[FileAndHash]()
	deletedfiles := Collections.NewList[FileAndHash]()
	nochange := Collections.NewList[FileAndHash]()

	log.Printf("Checking for deleted files...")

	tc := 1
	for k1, _ := range f1 {
		log.Printf("%d/%d", tc, total1)
		tc += 1
		_, ok := f2[k1]
		if !ok {
			deletedfiles.Add(FileAndHash{Path: k1, Hash: ""})
		}
	}

	log.Printf("Checking for changed and new files...")

	tc = 1
	for k2, v2 := range f2 {
		log.Printf("%d/%d", tc, total2)
		tc += 1

		v1, ok := f1[k2]
		if !ok {
			newfiles.Add(FileAndHash{Path: k2, Hash: ""})
		} else {
			checksum2, err2 := getFileChecksum(v2, md5.New())
			if err2 != nil {
				return
			}
			checksum1, err1 := getFileChecksum(v1, md5.New())
			if err1 != nil {
				return
			}
			if checksum1 == checksum2 {
				nochange.Add(FileAndHash{Path: k2, Hash: checksum2})
			} else {
				updatedfiles.Add(FileAndHash{Path: k2, Hash: checksum2})
			}
		}
	}

	SaveResult(newfiles, "new_files.txt")
	SaveResult(deletedfiles, "deleted_files.txt")
	SaveResult(updatedfiles, "updated_files.txt")
	println("-------------FolderDiff END-------------")

}

func SaveResult(files Collections.List[FileAndHash], outFileName string) {
	var output strings.Builder
	files.Sort(func(item1 FileAndHash, item2 FileAndHash) int {
		return strings.Compare(item1.Path, item2.Path)
	})
	for k := range files.GetSeq() {
		output.WriteString(k.Path + "\n")
	}
	fi := FileSystem.FileInfo_New(outFileName)
	err := fi.Create()
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = fi.WriteAllText(output.String())
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func getFiles(path string) (map[string]string, int, error) {
	res := map[string]string{}
	di := FileSystem.DirectoryInfo_New(path)

	fs, err := di.GetAllFiles()
	if err != nil {
		return nil, 0, err
	}
	fs.ForEachIndexed(func(item string, index int) {
		cpath, _ := strings.CutPrefix(item, path)
		res[cpath] = item
	})
	return res, fs.Size(), err
}

func getFileChecksum(filePath string, hash hash.Hash) (string, error) {
	//println(filePath)
	defer timeTrack(time.Now(), fmt.Sprintf("getFileChecksum [%s]: ", filePath))
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashstr := hex.EncodeToString(hash.Sum(nil))
	//println(hashstr)
	return hashstr, nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
