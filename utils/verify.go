package utils

import (
	"errors"

	"git.timschuster.info/rls.moe/catgi/config"
	"github.com/hlandau/passlib"
	"github.com/mishudark/dropbox-password"
)

// VerifyPassword takes a username, a password and the application
// configuration and validates the authentication.
func VerifyPassword(user, pass string, conf config.Configuration) error {
	for _, v := range conf.Users {
		if v.Username == user {
			switch v.AuthType {
			case config.ATPasslib:
				_, err := passlib.Verify(pass, v.PassHash)
				return err
			case config.ATDropbox:
				if len(conf.Pepper) != 32 {
					return errors.New("Pepper corrupt or missing")
				}
				isValid := password.IsValid(pass, v.PassHash, conf.Pepper)
				if isValid {
					return nil
				}
				return errors.New("Password Invalid")
			default:
				return errors.New("Invalid AT")
			}
		}
	}
	return errors.New("User not found")
}
