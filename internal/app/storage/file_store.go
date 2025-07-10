package storage

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
)

type FileStore struct {
	readFile  *os.File
	writeFile *os.File

	fileMu sync.Mutex

	scanner *bufio.Scanner
}

func LoadFile(fileStoragePath string) (*FileStore, error) {
	readFile, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	writeFile, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStore{
		readFile:  readFile,
		writeFile: writeFile,

		scanner: bufio.NewScanner(readFile),
	}, nil
}

//func (f *FileStore) Read() (map[string]string, error) {
//	if !f.scanner.Scan() {
//		if err := f.scanner.Err(); err != nil {
//			return nil, err
//		}
//
//		return nil, io.EOF
//	}
//
//	data := f.scanner.Bytes()
//
//	var deserialized map[string]string
//	if err := json.Unmarshal(data, &deserialized); err != nil {
//		return nil, err
//	}
//
//	return deserialized, nil
//}

func (f *FileStore) ReadRaw() ([]byte, error) {
	if !f.scanner.Scan() {
		if err := f.scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	raw := append([]byte(nil), f.scanner.Bytes()...)
	return raw, nil
}

func (f *FileStore) Append(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	b = append(b, '\n')

	f.fileMu.Lock()
	defer f.fileMu.Unlock()
	_, err = f.writeFile.Write(b)
	return err
}
