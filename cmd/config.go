package cmd

import (
	"fmt"
	"os"

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
		printError("cannot find rootToken in your .backend.yml config file")
		fmt.Println("\nMake sure to get your root token and save it in a .backend.yml config file.")
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

func getAuthToken() (tok string, ok bool) {
	tok = viper.GetString("authToken")
	if len(tok) == 0 {
		printError("cannot find authToken in your .backend.yml config file")
		fmt.Println("\nPlease run \"backend login\" to set up your credentials.")
		return
	}

	if _, err := backend.Me(tok); err == nil {
		ok = true
		return
	}

	// token expired/invalid, try to refresh
	email := viper.GetString("email")
	password := viper.GetString("password")

	newTok, err := backend.Login(email, password)
	if err != nil {
		printError("your auth token is invalid and could not be refreshed")
		fmt.Println("\nPlease run \"backend login\" again to set up your credentials.")
		return
	}

	if err := updateAuthToken(newTok); err != nil {
		printWarning("could not persist refreshed auth token: %v", err)
	}

	tok = newTok
	ok = true
	return
}

func updateAuthToken(newTok string) error {
	pk := viper.GetString("pubKey")
	region := viper.GetString("region")
	rtoken := viper.GetString("rootToken")
	email := viper.GetString("email")
	password := viper.GetString("password")

	s := fmt.Sprintf("pubKey: %s\nregion: %s\nrootToken: %s\nemail: %s\npassword: %s\nauthToken: %s", pk, region, rtoken, email, password, newTok)

	path := viper.ConfigFileUsed()
	if path == "" {
		path = "./.backend.yml"
	}

	return os.WriteFile(path, []byte(s), 0660)
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
