package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	staticbackend "github.com/staticbackendhq/core"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type ctxvalue int

const (
	ctxStatus ctxvalue = iota
	ctxStart
	ctxPath
)

var (
	verbose bool
)

func init() {
	// initialize minimum env variables
	//TODO: find another way to provide config to server
	os.Setenv("SB_FROM_CLI", "yes")
	os.Setenv("APP_ENV", "dev")
	os.Setenv("DATA_STORE", "mem")
	os.Setenv("JWT_SECRET", "fromcli")
	os.Setenv("MAIL_PROVIDER", "dev")
	os.Setenv("STORAGE_PROVIDER", "local")
	os.Setenv("FROM_EMAIL", "you@cli.com")
	os.Setenv("FROM_NAME", "from cli")
	os.Setenv("LOCAL_STORAGE_URL", "http://localhost:8099")
}

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

%s
	`,
		clbold(clsecondary("StatickBackend development server")),
		clnote("https://staticbackend.com/cli"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("no-color").Value.String() == "true" {
			color.Disable()
		}

		verbose = cmd.Flag("no-log").Value.String() == "false"
		f := cmd.Flag("port")

		uri := fmt.Sprintf(
			"http://localhost:%s/account/init?email=admin@dev.com&mem=1",
			f.Value.String(),
		)
		go createCustomer(uri, f.Value.String())

		staticbackend.Start("mem", f.Value.String())
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
}

func createCustomer(uri, port string) {
	fmt.Printf("%s: %s\n\n", clsecondary("server started at"), clbold("http://localhost:"+port))

	time.Sleep(300 * time.Millisecond)

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

	fmt.Printf("%s CTRL+C %s\n\n",
		clsecondary("press"),
		clsecondary("to quit and close server"),
	)
}
