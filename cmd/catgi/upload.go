package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
)

type handlerServePost struct{}

func newHandlerServePost() http.Handler {
	return &handlerServePost{}
}

func (h *handlerServePost) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("postFile", r.Context())

	log.Debug("Getting snowflake")
	flake, err := snowflakes.NewSnowflake()
	if err != nil {
		log.Error("Could not obtain a snowflake: ", err)
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}
	log.Debug("Request Snowflake is ", flake)

	err = r.ParseMultipartForm(25 * 1024 * 1024)
	if err != nil {
		log.Warn("Could not read form")
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}

	var file types.File
	httpFile, hdr, err := r.FormFile("data")
	if err != nil {
		log.Warn("Could not read form file")
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}
	fileData, err := ioutil.ReadAll(httpFile)
	if err != nil {
		log.Warn("Could not read form file")
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}
	file.Data = fileData
	log.Infof("Read %d bytes of a file", len(file.Data))
	dAt, err := types.FromString(r.Form.Get("delete_at"))
	if err != nil {
		log.Warn("Could not read delete time: ", err)
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}
	file.DeleteAt = dAt
	file.FileExtension = filepath.Ext(hdr.Filename)
	file.ContentType = http.DetectContentType(file.Data)
	file.Flake = flake

	{
		ctx := r.Context()
		usr := ctx.Value("user")
		if usr != nil {
			if val, ok := usr.(string); ok {
				log.Debug("User Context, setting owner")
				file.User = val
			}
		}
	}

	// TODO Implement Public Gallery
	file.Public = false

	err = curBe.Upload(flake, &file, r.Context())
	if err != nil {
		log.Warn("Could not commit file to database")
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}

	http.Redirect(rw, r, "/f/"+file.Flake+"/"+hdr.Filename, 302)
}
