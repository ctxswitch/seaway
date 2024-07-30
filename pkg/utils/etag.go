package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
)

func ETag(file string) string {
	hasher := md5.New()
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln("open", err)
	}
	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}
