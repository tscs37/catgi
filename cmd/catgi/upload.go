package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"git.timschuster.info/rls.moe/catgi/backend/common"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"git.timschuster.info/rls.moe/catgi/utils"
)

type handlerServePost struct {
	backend common.Backend
}

func newHandlerServePost(b common.Backend) http.Handler {
	return &handlerServePost{
		backend: b,
	}
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

	var file common.File
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
	log.Debugf("Read %d bytes of a file", len(file.Data))
	dAt, err := common.FromString(r.Form.Get("delete_at"))
	if err != nil {
		log.Warn("Could not read delete time: ", err)
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	}
	file.DeleteAt = dAt
	file.CreatedAt = common.FromTime(time.Now().UTC())
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

	var disableRedirect = false
	{
		val := r.Form.Get("disable_redirect")
		disableRedirect = (val == "on")
	}

	// TODO Implement Public Gallery
	file.Public = false

	r = r.WithContext(utils.PutHTTPIntoContext(r, r.Context()))

	// <- BEGIN BACKEND INTERACTION ->
	err = h.backend.Upload(flake, &file, r.Context())
	// -> END BACKEND INTERACTION <-

	if err != nil && !common.IsHTTPOption(err) {
		log.Warn("Could not commit file to database")
		rw.WriteHeader(500)
		fmt.Fprintf(rw, "Error: %s", err)
		return
	} else if common.IsHTTPOption(err) {
		httpopt := err.(common.ErrorHTTPOptions)
		httpopt.PassOverHTTP(rw)
		if httpopt.WantsTakeover() {
			httpopt.HTTPTakeover(r, rw, r.Context())
			return
		}
	}

	if !disableRedirect {
		http.Redirect(rw, r, "/f/"+file.Flake+"/"+hdr.Filename, 302)
	} else {
		fmt.Fprint(rw, "/f/"+file.Flake+"/"+hdr.Filename)
	}
}
