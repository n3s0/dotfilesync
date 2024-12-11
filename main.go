package main

import (
    "fmt"
    "os"
    "io"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/n3s0/dotfilesync/dfs"
)

type DotFileConfig struct {
    SyncBaseDir string
    XdgBaseDir string
    SyncDirs []string
    DotFiles []string
    DotFilesIgnore []string
}

var cfgFile string
var syncBaseDir string
var dotFilePaths []string
var sync bool

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of dotfilesync",
  Long:  `Print the version number of dotfilesync`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("dotfilesync v0.1")
  },
}

var rootCmd = &cobra.Command{
    Use:   "dotfilesync",
    Short: "A command-line dotfile syncing tool.",
    Long: `A command-line dotfile syncing tool.`,
    Run: func(cmd *cobra.Command, args []string) {
        var config DotFileConfig

        err := viper.Unmarshal(&config)
        if err != nil {
            fmt.Println("Error:", err)
            os.Exit(1)
        }
        
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        syncBaseDir = fmt.Sprintf("%s/%s", home, config.SyncBaseDir)
        err = dfs.CreateDirIfNotExist(syncBaseDir)
        if err != nil {
            fmt.Println("Error:", err)
            os.Exit(1)
        }

        xdgBaseDir := fmt.Sprintf("%s/%s/%s", home, config.SyncBaseDir, config.XdgBaseDir)
        err = dfs.CreateDirIfNotExist(xdgBaseDir)
        if err != nil {
            fmt.Println("Error", err)
            os.Exit(1)
        }


        if sync {
            for _, dir := range config.SyncDirs {
                syncPath := fmt.Sprintf("%s%s", syncBaseDir, dir)
                fmt.Println("syncpath", syncPath)

                err := dfs.CreateDirIfNotExist(syncPath)
                if err != nil {
                    fmt.Println("Error:", err)
                    os.Exit(1)
                }
            }

            dotFilePaths = config.DotFiles

            for _, file := range dotFilePaths {
                syncFilePath := fmt.Sprintf("%s/%s", home, file)
                dotFilePath := fmt.Sprintf("%s/%s", syncBaseDir, file)
                fmt.Println("syncpath2", syncFilePath)
                fmt.Println("dotfilepath", dotFilePath)

                needsSynced, err := dfs.CompareFiles(syncFilePath, dotFilePath)
                if err != nil {
                    fmt.Println("Error:", err)
                    os.Exit(1)
                }

                switch needsSynced {
                case "dotfile":
                    fmt.Printf("Copying %s -> %s\n", dotFilePath, syncFilePath)
                    dfs.CopyFile(dotFilePath, syncFilePath)
                case "syncfile":
                    fmt.Printf("Copying %s -> %s\n", syncFilePath, dotFilePath)
                    dfs.CopyFile(syncFilePath, dotFilePath)
                case "synced":
                    fmt.Println("All files are synced")
                default:
                    fmt.Println("All files are synced")
                }
            }
        }
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println("Unable to execute. Error provided:", err)
        os.Exit(1)    
    }
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/dotfilesync/config.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&sync, "sync", "s", false, "sync dotfiles to dotfile directory")
}

func initConfig() {
    var configDir string = ".config/dotfilesync"

    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        defaultConfigPath := fmt.Sprintf("%s/%s", home, configDir)

        viper.AddConfigPath(defaultConfigPath)
        viper.SetConfigType("yaml")
        viper.SetConfigName("config.yaml")
    }

    if err := viper.ReadInConfig(); err != nil {
        fmt.Println("Cannot read config:", err)
        os.Exit(1)
    }
}

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

func main() {
    Execute()
}
