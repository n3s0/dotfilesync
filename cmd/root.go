package cmd

import (
    "fmt"
    "os"

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
