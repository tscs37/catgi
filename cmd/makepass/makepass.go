package main

import (
	"fmt"

	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/hlandau/passlib"
	"github.com/howeyc/gopass"
)

func main() {
	ctx := logger.NewLoggingContext()
	log := logger.LogFromCtx("makepass", ctx)
	user := &config.UserConfig{}
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
	dat, err := json.Marshal(user)
	if err != nil {
		log.Error("Could not marshal data to user: ", err)
		return
	}
	fmt.Print("--- USER  DATA ---\n\n")
	fmt.Printf("%s\n\n", dat)
	fmt.Print("--- END OF PRG ---\n")
}
