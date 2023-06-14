package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Version is the current version of the CLI
	Version = "v1.5.0"
)

var (
	clgreen     = color.FgGreen.Render
	clinfo      = color.Info.Render
	clnote      = color.Note.Render
	cllight     = color.Light.Render
	clerror     = color.Error.Render
	cldanger    = color.Danger.Render
	cldebug     = color.Debug.Render
	clnotice    = color.Notice.Render
	clsuccess   = color.Success.Render
	clcomment   = color.Comment.Render
	clprimary   = color.Primary.Render
	clwarning   = color.Warn.Render
	clquestion  = color.Question.Render
	clsecondary = color.Secondary.Render
	clbold      = color.Bold.Render
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "backend",
	Short: "StaticBackend CLI for local development, managing resources, and your account.",
	Long: fmt.Sprintf(`
%s

This CLI gives you the following functionalities:

- A local development server: %s
- Managing your backend resources: %s
- Managing your account: %s

Use "backend server" to start your local dev server.

Use "backend login --dev" to automatically configure for local dev.
	`,
		clbold(clsecondary("StaticBackend CLI "+Version)),
		clbold("backend server"),
		clsecondary("db, function, form, etc"),
		clsecondary("billing"),
	),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("version").Value.String() == "true" {
			fmt.Println(Version)
		} else {
			fmt.Println(cmd.Long)
			fmt.Println("")
			if err := cmd.Usage(); err != nil {
				fmt.Println(err)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/.backend.yaml)")
	rootCmd.PersistentFlags().Bool("no-color", false, "turns color off")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "display current version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// add current working directory
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		confDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(pwd)
		viper.AddConfigPath(path.Join(confDir, "backend"))
		viper.SetConfigName(".backend")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("===\n%s: %v\n===\n\n", clwarning("No config file used"), err)
	}
}
