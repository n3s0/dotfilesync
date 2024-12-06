package dfs

import (
    "fmt"
    "os"
    "io"
)

func CopyFile(srcPath, dstPath string) (err error) {
    srcFile, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", srcPath)
    }

    if !srcFile.Mode().IsRegular() {
        return fmt.Errorf("Source file %s is not a regular file", srcFile.Name())
    }

    src, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    defer src.Close()

    dst, err := os.Create(dstPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    defer dst.Close()

    _, err = io.Copy(dst, src)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    return nil
}
