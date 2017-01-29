package common

import (
	"context"

	"reflect"

	"fmt"

	"git.timschuster.info/rls.moe/catgi/logger"
	"github.com/mitchellh/mapstructure"
)

// isPtrInterface will return false if the given interface is not
// a pointer
// This function is magic, do not touch.
func isPtrInterface(ptr interface{}, ctx context.Context) bool {
	log := logger.LogFromCtx("common.isPtrInterface", ctx)
	ptrVal := reflect.ValueOf(ptr)
	log.Debugf("Type of Ptr: '%s'", ptrVal.Kind())
	if ptrVal.Kind() != reflect.Ptr {
		log.Error("Ptr was not a pointer, slap programmer.")
		return false
	}
	return true
}

// DecodeOption is a struct containing a PreCheckHook and a PostDecodeHook
type DecodeOption struct {
	PreCheckHook func(map[string]interface{}) error
}

// ConfigMustHave checks if a top-level key was set
// in the configuration.
//
// It checks not if the key is simple set to an empty string
// or other nil values.
func ConfigMustHave(options ...string) DecodeOption {
	return DecodeOption{
		PreCheckHook: func(conf map[string]interface{}) error {
			for _, option := range options {
				if _, ok := conf[option]; !ok {
					return fmt.Errorf("Key %s not found but required", option)
				}
			}
			return nil
		},
	}
}

// ConfigDefault is a DecodeOption that sets a Default
// value to a top-level key if it's not specified.
func ConfigDefault(option string, value interface{}) DecodeOption {
	return DecodeOption{
		PreCheckHook: func(conf map[string]interface{}) error {
			if _, ok := conf[option]; !ok {
				conf[option] = value
			}
			return nil
		},
	}
}

// DecodeConfig will decode the PRM map into the PTR parameter
// The map names must be specified int the "cgc" struct tag,
// for reference check the mapstructure library
func DecodeConfig(ptr interface{}, prm map[string]interface{}, ctx context.Context, opts ...DecodeOption) error {
	log := logger.LogFromCtx("common.DecodeConfig", ctx)

	isPtrInterface(ptr, ctx)

	log.Debug("Running PreCheckHook")

	for k := range opts {
		if opts[k].PreCheckHook != nil {
			err := opts[k].PreCheckHook(prm)
			if err != nil {
				return err
			}
		}
	}

	log.Debug("Create Default Config Decoder")

	var md mapstructure.Metadata
	conf := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		ZeroFields:       false,
		Result:           ptr,
		Metadata:         &md,
		TagName:          "cgc",
	}

	log.Debug("Create Decoder")
	decoder, err := mapstructure.NewDecoder(conf)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("Decoding Map...")
	err = decoder.Decode(prm)
	if err != nil {
		log.Error(err)
		return err
	}

	log.
		WithField("decoded_keys", md.Keys).
		WithField("config_result", prm).
		Debug("Decoded Config")
	return nil
}
