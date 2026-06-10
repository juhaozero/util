package common

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("存在同名文件")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return err
}

func FileMove(src string, dst string) (err error) {
	if dst == "" {
		return nil
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	revoke := false
	dir := filepath.Dir(dst)
Redirect:
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
		if !revoke {
			revoke = true
			goto Redirect
		}
	}
	return os.Rename(src, dst)
}

func DeLFile(filePath string) error {
	return os.RemoveAll(filePath)
}

func TrimSpace(target interface{}) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return
	}
	t = t.Elem()
	v := reflect.ValueOf(target).Elem()
	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(strings.TrimSpace(v.Field(i).String()))
		}
	}
}

// FileExist 判断文件是否存在
func FileExist(path string) bool {
	fi, err := os.Lstat(path)
	if err == nil {
		return !fi.IsDir()
	}
	return !os.IsNotExist(err)
}
