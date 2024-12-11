package dfs

import (
    "fmt"
    "os"
    "io"
)

/*
This will copy files over. We will use this to copy configuration files to the 
sync path. But, this can also be used to copy regular text files pretty anywhere.
*/
func CopyFile(srcPath, dstPath string) (e error) {
    srcFile, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
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

func CopyDir(srcPath, dstPath string) (e error) {
    srcDir, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    if !srcDir.IsDir() {
        return fmt.Errorf("Error: Path (%s) does not point to directory\n", srcDir)
    }

    files, err := os.ReadDir(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    fmt.Println(files)

    return nil
}

/*
Will create a directory in the desired path with mode 0755 if it doesn't already
exist. This will be useful for when we start scaffolding sync directory structures
for copying files.
*/
func CreateDirIfNotExist(path string) (e error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        err := os.Mkdir(path, 0755)
        if err != nil {
            return fmt.Errorf("Error: %v", err)
        }
    }
    return nil
}

/*
Need a function that compares the sync path to the dotfile path.

Some ideas I'm throwing around is checking the modified time of both files
and their hashes to confirm they aren't different.

This will either return errors or dotfiles / sync. 

For now for simplicity. I'll just compare the modified time and make a decision
based on that.
*/

func CompareFiles(syncPath, dotFilePath string) (sync string, e error) {
    const (
        dotFile string = "dotfile"
        syncFile string = "syncfile"
        synced string = "synced"
    )

    syncFileInfo, syncErr := os.Stat(syncPath)
    if syncErr != nil {
        return "", fmt.Errorf("Error: %v", syncErr)
    }

    dotFileInfo, dotErr := os.Stat(dotFilePath)
    if dotErr != nil {
        return "", fmt.Errorf("Error: %v", dotErr)
    }

    if syncFileInfo.ModTime().Before(dotFileInfo.ModTime()) {
        return syncFile, nil
    } else if dotFileInfo.ModTime().Before(syncFileInfo.ModTime()) {
        if os.IsNotExist(syncErr) {
            return syncFile, nil
        }
        return dotFile, nil
    } else {
        return synced, nil
    }
}
