package main

import (
    "fmt"
    "os"
    "io"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

type Config struct {
	Sync struct {
		SyncDir string `mapstructure:"sync_dir"`
	} `mapstructure:"sync"`
	Dotfiles struct {
		DirPaths  []string `mapstructure:"dir_paths"`
		FilePaths []string `mapstructure:"file_paths"`
	} `mapstructure:"dotfiles"`
}

var cfgFile string
var syncBaseDir string
var dotFilePaths []string
var verbose bool
var sync bool
var backup bool
var push bool

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
        var config Config

        err := viper.Unmarshal(&config)
        if err != nil {
            fmt.Printf("Error: %s\n", err)
            os.Exit(1)
        }

        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Printf("Error: %s\n", err)
            os.Exit(1)
        }

        // Default directory for putting the sync directory is the root of home.
        syncDir := fmt.Sprintf("%s/%s", home, config.Sync.SyncDir)
        err = CreateDirIfNotExist(syncDir)
        if err != nil {
            fmt.Println("Error", err)
            os.Exit(1)
        }

        dotDirPaths := config.Dotfiles.DirPaths
        dotFilePaths = config.Dotfiles.FilePaths

        if backup {
            fmt.Printf("Backing up directories\n")
            
            for _, dir := range dotDirPaths {
                syncDirPath := fmt.Sprintf("%s/%s", home, dir)
                dotDirPath := fmt.Sprintf("%s/%s", syncDir, dir)

                
                err := CopyDir(syncDirPath, dotDirPath)
                if err != nil {
                    fmt.Printf("Error: %s\n", err)
                }
            } 
            
            fmt.Printf("Backing up dotfiles\n")
            
            for _, file := range dotFilePaths {
                syncFilePath := fmt.Sprintf("%s/%s", home, file)
                dotFilePath := fmt.Sprintf("%s/%s", syncDir, file)

                err := CopyFile(syncFilePath, dotFilePath)
                if err != nil {
                    fmt.Printf("Error: %s\n", err)
                }
            }
        }

        if sync {
            fmt.Println("I will eventually sync the files to the home directory")
        }

        if push {
            fmt.Println("I will eventually commit and push the dotfiles git repo")
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

    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.config/dotfilesync/config.yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "Show verbose output")
    rootCmd.PersistentFlags().BoolVarP(&sync, "sync", "s", false, "Sync dotfiles to the home directory. Will sync to dotfiles directory too.")
    rootCmd.PersistentFlags().BoolVarP(&backup, "backup", "b", false, "Backup dotfiles to sync directory")
    rootCmd.PersistentFlags().BoolVarP(&push, "push", "p", false, "Adding, commit, and push files to git repo")
}

func initConfig() {
    // Base directory for the configuration is $HOME/.config/dotfilesync. Config
    // flag will update this.
    var configDir string = ".config/dotfilesync"

    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Printf("Error: %s", err)
            os.Exit(1)
        }

        defaultConfigPath := fmt.Sprintf("%s/%s", home, configDir)

        viper.AddConfigPath(defaultConfigPath)
        viper.SetConfigType("yaml")
        viper.SetConfigName("config.yaml")
    }

    if err := viper.ReadInConfig(); err != nil {
        fmt.Println("Cannot read config:", err)
        os.Exit(1)
    }
    
    fmt.Printf("Config file loaded: %s\n", viper.ConfigFileUsed())
}

/*
This will copy files over. We will use this to copy configuration files to the 
sync path. But, this can also be used to copy regular text files pretty anywhere.
*/
func CopyFile(srcPath, dstPath string) (e error) {
    srcFile, err := os.Stat(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v\n", err)
    }

    if !srcFile.Mode().IsRegular() {
        return fmt.Errorf("Source file %s is not a regular file", srcFile.Name())
    }

    src, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v\n", err)
    }

    defer src.Close()

    dst, err := os.Create(dstPath)
    if err != nil {
        return fmt.Errorf("Error: %v\n", err)
    }

    defer dst.Close()

    if verbose {
        fmt.Printf("Copying %s -> %s\n", src.Name(), dst.Name())
    }
    
    _, err = io.Copy(dst, src)
    if err != nil {
        return fmt.Errorf("Error: %s\n", err)
    }

    return nil
}

func CopyDir(srcPath, dstPath string) (e error) {
    err := os.MkdirAll(dstPath, 0755)
    if err != nil {
        return fmt.Errorf("Error: %s", err)
    }

    files, err := os.ReadDir(srcPath)
    if err != nil {
        return fmt.Errorf("Error: %v", err)
    }

    for _, file := range files {
        src := fmt.Sprintf("%s/%s", srcPath, file.Name())
        dst := fmt.Sprintf("%s/%s", dstPath, file.Name())

        if verbose {
            fmt.Printf("Copying %s -> %s\n", src, dst)
        }

        if file.IsDir() {
            if file.Name() == ".git" {
                if verbose {
                    fmt.Println("Skipping git directory")
                }
            } else {
                err = CopyDir(src, dst)
                if err != nil {
                    return fmt.Errorf("Error: %s\n", err)
                }
            }
        } else {
            err = CopyFile(src, dst)
            if err != nil {
                return fmt.Errorf("Error: %s\n", err)
            }
        }
    }

    return nil
}

/*
Will create a directory in the desired path with mode 0755 if it doesn't already
exist. This will be useful for when we start scaffolding sync directory structures
for copying files.
*/
func CreateDirIfNotExist(path string) (e error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        fmt.Printf("Directory path %s does not exist. Creating it.\n", path)
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

func CompareDirs(syncPath, dotFilePath string) (sync string, e error) {
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

func GitAdd(repoPath string) (e error) {

    return nil
}

func GitCommit(repoPath string) (e error) {
    return nil
}

func GitPush(repoPath string) (e error) {
    return nil
}

func main() {
    Execute()
}
