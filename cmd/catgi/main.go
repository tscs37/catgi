package main

import (
	"errors"
	"net/http"

	"encoding/json"

	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/InVisionApp/rye"
	"github.com/gorilla/mux"
)

var (
	curBe backend.Backend
)

func main() {
	ctx := logger.NewLoggingContext()
	logger.SetLoggingLevel("info", ctx)

	log := logger.LogFromCtx("main", ctx)

	conf, err := config.LoadConfig("./conf.json")

	log.Info("Starting Backend")
	be, err := backend.NewBackend(conf.Backend.Name, conf.Backend.Params, ctx)
	if err != nil {
		log.Errorf("Error: %s", err)
		return
	}
	curBe = be
	log.Infof("Loaded '%s' Backend Driver", be.Name())
	mwHandler := rye.NewMWHandler(rye.Config{})

	log.Info("Beginning Cleanup")

	err = be.CleanUp(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Finished Cleanup")

	router := mux.NewRouter()
	router.Handle("/file", mwHandler.Handle([]rye.Handler{
		injectLogToRequest,
		serveGet,
	})).Methods("GET")

	router.Handle("/file", mwHandler.Handle([]rye.Handler{
		injectLogToRequest,
		servePost,
	}))

	router.Handle("/", mwHandler.Handle([]rye.Handler{
		injectLogToRequest,
		serveSite,
	}))

	router.Handle("/login", mwHandler.Handle([]rye.Handler{
		injectLogToRequest,
		serveLogin,
	}))

	router.Handle("/auth", mwHandler.Handle([]rye.Handler{
		injectLogToRequest,
		serveAuth,
	}))

	http.ListenAndServe("[::1]:8080", router)
}

func injectLogToRequest(_ http.ResponseWriter, r *http.Request) *rye.Response {
	return &rye.Response{
		Context: logger.InjectLogToContext(r.Context()),
	}
}

func serveAuth(rw http.ResponseWriter, r *http.Request) *rye.Response {
	return nil
}

func serveLogin(rw http.ResponseWriter, r *http.Request) *rye.Response {
	return nil
}

func serveSite(rw http.ResponseWriter, r *http.Request) *rye.Response {
	return nil
}

func servePost(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log := logger.LogFromCtx("postFile", r.Context())

	log.Debug("Getting snowflake")
	flake, err := snowflakes.NewSnowflake()
	if err != nil {
		log.Error("Could not obtain a snowflake: ", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	log.Debug("Request Snowflake is ", flake)

	log.Debug("Parsing Form")
	err = r.ParseForm()
	if err != nil {
		log.Warn("Form not parse, request aborted")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	dat := r.PostForm.Get("file")
	var file types.File
	err = json.Unmarshal([]byte(dat), &file)
	if err != nil {
		log.Warn("Could not parse incoming file")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	file.ContentType = http.DetectContentType(file.Data)
	file.Flake = flake

	// TODO Implement Public Gallery
	file.Public = false

	err = curBe.Upload(flake, &file, r.Context())
	if err != nil {
		log.Warn("Could not commit file to database")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	// Clean file part and return and (empty) file JSON document
	file.Data = []byte{}

	{
		log.Debug("Parsing data to JSON")
		dat, err := json.Marshal(file)
		if err != nil {
			log.Warn("Error while parsing data to json: ", err)
			return &rye.Response{
				Err:           err,
				StopExecution: true,
			}
		}

		log.Debug("Writing out response")

		rye.WriteJSONResponse(rw, 200, dat)

		return nil
	}
}

func serveGet(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log := logger.LogFromCtx("getFile", r.Context())

	log.Debug("Parsing Form")
	err := r.ParseForm()
	if err != nil {
		log.Warn("Form not parsed, request aborted")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Loading Flake from Form")
	flake := r.Form.Get("flake")
	if len(flake) == 0 {
		log.Warn("Form contained no flake")
		return &rye.Response{
			Err:           errors.New("Need flake to find file"),
			StopExecution: true,
		}
	}

	log.Debug("Loading File from Backend")
	f, err := curBe.Get(flake, r.Context())
	if err != nil {
		log.Warn("File error on backend: ", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Parsing data to JSON")
	dat, err := json.Marshal(f)
	if err != nil {
		log.Warn("Error while parsing data to json: ", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Writing out response")

	rye.WriteJSONResponse(rw, 200, dat)

	return nil
}
