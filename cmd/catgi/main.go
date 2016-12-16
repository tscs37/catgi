package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"mime"

	"path/filepath"

	"bitbucket.org/taruti/mimemagic"
	"git.timschuster.info/rls.moe/catgi/backend"
	_ "git.timschuster.info/rls.moe/catgi/backend/b2"
	_ "git.timschuster.info/rls.moe/catgi/backend/buntdb"
	"git.timschuster.info/rls.moe/catgi/backend/types"
	"git.timschuster.info/rls.moe/catgi/config"
	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
	"github.com/InVisionApp/rye"
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
)

var (
	curBe backend.Backend
)

func main() {
	ctx := logger.NewLoggingContext()
	logger.SetLoggingLevel("debug", ctx)

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

	h := hashids.NewData()
	h.MinLength = 1
	h.Salt = "catgi.rls.moe"
	hd := hashids.NewWithData(h)
	fmt.Printf("%s\n", hd.Decode("Zo8KDWGBzkKbQ"))

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

	log.Info("Starting HTTP Service")

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
	log := logger.LogFromCtx("postFile", r.Context())
	dat, err := ioutil.ReadFile("./index.html")
	if err != nil {
		log.Error("Could not load file from disk: ", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	rw.WriteHeader(200)
	rw.Header().Add("Content-Type", "application/html")
	rw.Write(dat)
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

	err = r.ParseMultipartForm(25 * 1024 * 1024)
	if err != nil {
		log.Warn("Could not read form")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	var file types.File
	httpFile, hdr, err := r.FormFile("data")
	if err != nil {
		log.Warn("Could not read form file")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	fileData, err := ioutil.ReadAll(httpFile)
	if err != nil {
		log.Warn("Could not read form file")
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	file.Data = fileData
	log.Infof("Read %d bytes of a file", len(file.Data))
	dAt, err := types.FromString(r.Form.Get("delete_at"))
	if err != nil {
		log.Warn("Could not read delete time: ", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}
	file.DeleteAt = dAt
	file.FileExtension = filepath.Ext(hdr.Filename)
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

	{
		http.Redirect(rw, r, "/file?flake="+file.Flake, 302)

		return nil
	}
}

func serveGet(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log := logger.LogFromCtx("getFile", r.Context())

	log.Debug("Parsing Form")
	err := r.ParseForm()
	if err != nil {
		log.Warn("Form not parsed, request aborted")
		rye.WriteJSONStatus(rw, "error", err.Error(), 500)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Loading Flake from Form")
	flake := r.Form.Get("flake")
	if len(flake) == 0 {
		log.Warn("Form contained no flake")
		err = errors.New("Missing Flake Parameter")
		rye.WriteJSONStatus(rw, "error", err.Error(), 500)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Loading File from Backend")
	f, err := curBe.Get(flake, r.Context())
	if err != nil {
		log.Warn("File error on backend: ", err)
		rye.WriteJSONStatus(rw, "error", err.Error(), 500)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	log.Debug("Writing out response")

	mimetype := ""

	if f.ContentType == "" {
		if len(f.Data) > 1024 {
			mimetype = mimemagic.Match("image/png", f.Data[:1024])
		} else {
			mimetype = mimemagic.Match("image/png", f.Data)
		}
	} else {
		mimetype = f.ContentType
	}

	ext := ""

	{
		exts, err := mime.ExtensionsByType(mimetype)
		if err != nil || len(exts) < 1 {
			log.Warn("MIMEType without extension")
		} else {
			ext = exts[0]
		}
	}

	rw.Header().Add("Content-Disposition", "inline; filename="+f.Flake+ext)
	rw.Header().Add("Content-Type", mimetype)
	_, err = rw.Write(f.Data)
	if err != nil {
		log.Errorf("Error on store: %s", err)
		return &rye.Response{
			Err:           err,
			StopExecution: true,
		}
	}

	return nil
}
