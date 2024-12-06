package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/n3s0/dotfilesync/dfs"
)

var cfgFile string
var syncDir string
var sync bool

var rootCmd = &cobra.Command{
    Use:   "dotfilesync",
    Short: "A command-line dotfile syncer.",
    Long: `A command-line dotfile syncer.`,
    Run: func(cmd *cobra.Command, args []string) {
        if sync {
            for  
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
