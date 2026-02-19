package cmd

import (
	"fmt"
	"net/http"
	"time"

	staticbackend "github.com/staticbackendhq/core"
	sbconfig "github.com/staticbackendhq/core/config"
	"github.com/staticbackendhq/core/logger"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a development server.",
	Long: fmt.Sprintf(`
%s

You may develop your application locally using the development server.

It has a direct mapping with StaticBackend API. You'll need no code changes
when going from local to production.

There are some limitations that you can learn more about here.

https://staticbackend.dev/cli
	`,
		clbold("StaticBackend development server"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		f := cmd.Flag("port")
		persistData := cmd.Flag("persist-data").Value.String() == "true"

		uri := fmt.Sprintf(
			"http://localhost:%s/account/init?email=admin@dev.com&mem=1",
			f.Value.String(),
		)

		go createCustomer(uri, f.Value.String())

		c := sbconfig.AppConfig{
			AppEnv:          "dev",
			FromCLI:         "yes",
			Port:            f.Value.String(),
			DatabaseURL:     "mem",
			DataStore:       "mem",
			RedisHost:       "mem",
			LocalStorageURL: "http://localhost:8099",
			ActivateFlag:    "no-stripe-test-flag",
		}

		if persistData {
			c.DatabaseURL = "local.db"
			c.DataStore = "sqlite"
		}

		log := logger.Get(c)

		staticbackend.Start(c, log)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serverCmd.Flags().Int32P("port", "p", 8099, "dev server port")
	serverCmd.Flags().Bool("no-log", false, "prevents printing requests/responses info")
	serverCmd.Flags().Bool("persist-data", false, "persists data across usage")
}

func createCustomer(uri, port string) {
	fmt.Printf("server started at: %s\n\n", clbold("http://localhost:"+port))

	time.Sleep(500 * time.Millisecond)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("press CTRL+C to quit and close server\n\n")
}
