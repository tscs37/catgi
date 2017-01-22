package main

import (
	"fmt"

	"encoding/json"

	"os"

	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/hlandau/passlib"
	"github.com/howeyc/gopass"
	"github.com/mishudark/dropbox-password"
)

func main() {
	ctx := logger.NewLoggingContext()
	log := logger.LogFromCtx("makepass", ctx)
	user := &config.UserConfig{}
	if pepper := os.Getenv("CATGI_PEPPER"); len(pepper) == 32 {
		fmt.Print("--- DROPBOX MD ---\n")
		fmt.Print("--- INPUT USER ---\n")
		fmt.Printf("Pepper  : %s\n", pepper)
		fmt.Print("Username: ")
		fmt.Scanln(&(user.Username))
		fmt.Print("Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			log.Error("Could not read password from stdin: ", err)
			return
		}
		fmt.Print("--- PROCESSING ---\n")
		hash, err := password.Hash(string(pass), pepper)
		if err != nil {
			log.Error("Could not hash password: ", err)
			return
		}
		user.PassHash = hash
		user.AuthType = config.ATDropbox
	} else if len(pepper) > 0 && len(pepper) != 32 {
		fmt.Print("--- DROPBOX MD ---\n")
		fmt.Print("Pepper is not 32 characters, aborting.")
		return
	} else {
		fmt.Print("--- LEGACY  MD ---\n")
		printLegacyBanner()
		fmt.Print("--- INPUT USER ---\n")
		fmt.Print("Username: ")
		fmt.Scanln(&(user.Username))
		fmt.Print("Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			log.Error("Could not read password from stdin: ", err)
			return
		}
		fmt.Print("--- PROCESSING ---\n")
		hash, err := passlib.Hash(string(pass))
		if err != nil {
			log.Error("Could not hash password: ", err)
			return
		}
		user.PassHash = hash
		user.AuthType = config.ATPasslib
	}
	dat, err := json.Marshal(user)
	if err != nil {
		log.Error("Could not marshal data to user: ", err)
		return
	}
	fmt.Print("--- USER  DATA ---\n\n")
	fmt.Printf("%s\n\n", dat)
	fmt.Print("--- END OF PRG ---\n")
}

func printLegacyBanner() {
	fmt.Println("makepass is running in legacy mode. You can")
	fmt.Println("use legacy mode to continue using old versions")
	fmt.Println("of catgi. It is recommended you use the")
	fmt.Println("CATGI_PEPPER environment variable to set")
	fmt.Println("a pepper for password encryption.")
}
