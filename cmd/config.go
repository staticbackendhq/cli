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

	"github.com/spf13/viper"
	"github.com/staticbackendhq/backend-go"
)

func getPublicKey() (pubKey string, ok bool) {
	pubKey = viper.GetString("pubKey")
	if len(pubKey) == 0 {
		fmt.Printf("%s\n", cldanger("cannot find pubKey in your .backend config file"))
		fmt.Println("\nMake sure to get your StaticBackend public key and save it in a .backend YAML config file.")
		fmt.Println("\nFor instance:")
		fmt.Printf("\n\t%s: na1", clsecondary("region"))
		fmt.Printf("\n\t%s: your-key-here", clsecondary("pubKey"))
		fmt.Println("\nYou received your public key when you created your account via email.")
		return
	}

	ok = true
	return
}

func getRootToken() (tok string, ok bool) {
	tok = viper.GetString("rootToken")
	if len(tok) == 0 {
		fmt.Printf("%s\n", cldanger("cannot find rootToken in your .backend config file"))
		fmt.Println("\nMake sure to get your root token and save it in a .backend YAML config file.")
		fmt.Println("\nFor instance:")
		fmt.Printf("\n\t%s: na1", clsecondary("region"))
		fmt.Printf("\n\t%s: your-key-here", clsecondary("pubKey"))
		fmt.Printf("\n\t%s: your-root-token-here", clsecondary("rootToken"))
		fmt.Println("\nYou received your root token when you created your account via email.")
		return
	}

	ok = true
	return
}

func setBackend() bool {
	pk, ok := getPublicKey()
	if !ok {
		return false
	}

	backend.PublicKey = pk

	region := viper.GetString("region")
	if len(region) == 0 {
		region = "na1"
	}

	backend.Region = region

	return true
}
