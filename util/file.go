package util

import (
	"os"
	"path/filepath"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 创建目录及必要的父目录
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func WriteBase64File(base64Str string, dst string) error {
	base64, err := DecodeFileBase64(base64Str)
	if err != nil {
		return err
	}
	return WriteFile(base64, dst)
}
func WriteFile(bytes []byte, dst string) error {
	// 创建目标文件所在的目录
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	// 创建文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// 确保文件关闭，并检查关闭时的错误
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	// 将字节数据写入文件
	_, err = out.Write(bytes)
	if err != nil {
		return err
	}
	// 确保数据被刷新到磁盘
	if err := out.Sync(); err != nil {
		return err
	}
	return nil
}
