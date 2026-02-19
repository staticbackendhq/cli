package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyTarget string

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Proxy requests to production instead of using the local dev server.",
	Long: fmt.Sprintf(`
%s

This command starts a proxy server relaying your requests to the production
backend.

It is useful if you'd like to test your application against the production
backend without deploying your application changes yet.

All requests are proxy as-is, and the responses sent to you without any
modifications.
	`, clbold("Proxy requests to production")),
	Run: func(cmd *cobra.Command, args []string) {
		f := cmd.Flag("port")
		startProxy(f.Value.String())
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// proxyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	proxyCmd.Flags().Int32P("port", "p", 8099, "port for the proxy server")
}

func startProxy(port string) {
	region := viper.GetString("region")
	if len(region) == 0 {
		printError("Missing a region config entry in your config file")
		os.Exit(1)
	}

	proxyTarget = fmt.Sprintf("https://%s.staticbackend.dev", region)

	http.HandleFunc("/", proxy)

	fmt.Printf("Proxy server started at: http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	t, err := url.Parse(proxyTarget)
	if err != nil {
		log.Fatal("invalid proxy target: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(t)

	// Update the headers to allow for SSL redirection
	r.URL.Host = "na1.staticbackend.dev"
	r.URL.Scheme = "https"
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = "na1.staticbackend.dev"

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)
}
