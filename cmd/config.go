package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/staticbackendhq/backend-go"
)

func getPublicKey() (pubKey string, ok bool) {
	pubKey = viper.GetString("pubKey")
	if len(pubKey) == 0 {
		printError("cannot find pubKey in your .backend.yml config file")
		fmt.Println("\nMake sure to get your StaticBackend public key and save it in a .backend.yml YAML config file.")
		fmt.Println("\nFor instance:")
		fmt.Printf("\n\tregion: na1")
		fmt.Printf("\n\tpubKey: your-key-here")
		fmt.Println("\nYou received your public key when you created your account via email.")
		fmt.Printf("\n%s", clbold("use \"backend login --dev\" to work with the development server.\n\n"))
		return
	}

	ok = true
	return
}

func getRootToken() (tok string, ok bool) {
	tok = viper.GetString("rootToken")
	if len(tok) == 0 {
		printError("cannot find rootToken in your .backend config file")
		fmt.Println("\nMake sure to get your root token and save it in a .backend YAML config file.")
		fmt.Println("\nFor instance:")
		fmt.Printf("\n\tregion: na1")
		fmt.Printf("\n\tpubKey: your-key-here")
		fmt.Printf("\n\trootToken: your-root-token-here")
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
		region = "dev"
	}

	backend.Region = region

	return true
}
