/*
Copyright © 2020 Focus Centric inc. <dominicstpierre@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
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
	`, clbold(clsecondary("Proxy requests to production"))),
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
		fmt.Printf("%s %s %s\n", cldanger("Missing a"), clerror("region"), cldanger("config entry in your config file"))
		os.Exit(1)
	}

	proxyTarget = fmt.Sprintf("https://%s.staticbackend.com", region)

	http.HandleFunc("/", proxy)

	fmt.Printf("%s http://localhost:%s\n", clsecondary("Proxy server started at:"), port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func proxy(w http.ResponseWriter, r *http.Request) {
	t, err := url.Parse(proxyTarget)
	if err != nil {
		log.Fatal("invalid proxy target: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(t)

	// Update the headers to allow for SSL redirection
	r.URL.Host = "na1.staticbackend.com"
	r.URL.Scheme = "https"
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = "na1.staticbackend.com"

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(w, r)
}
