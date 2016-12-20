package main

import (
	"fmt"
	"net/http"

	"net/url"

	"time"

	"git.timschuster.info/rls.moe/catgi/logger"
	"git.timschuster.info/rls.moe/catgi/snowflakes"
)

// This handler reports clicks to Piwik is a bit useless though :/
type handlerPiwik struct {
	piwikBase   string
	piwikId     string
	ignoreError bool
	enabled     bool
	next        http.Handler
}

func newHandlerPiwik(base, id string, enable, ignoreErr bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &handlerPiwik{
			piwikId:     id,
			piwikBase:   base,
			ignoreError: ignoreErr,
			enabled:     enable,
			next:        next,
		}
	}
}

var client = &http.Client{
	Timeout: 2 * time.Second,
}

func (h *handlerPiwik) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.LogFromCtx("piwik", r.Context())

	// This is only backend tracking, I do no belief that DNT is necessary
	// but correct me if I'm wrong. For now I'm going to use it to see
	// traffic statistics.
	if h.enabled /*&& !(r.Header.Get("DNT") == "1")*/ {

		piwikCall := "https://%s"

		snowflake, err := snowflakes.NewSnowflake()
		if err != nil {
			snowflake = "randerr"
		}
		piwikCall = fmt.Sprintf(piwikCall, h.piwikBase)

		log.Debug("Piwik Call is ", piwikCall)

		log.Debug("Starting call...")

		go func() {
			resp, err := client.PostForm(piwikCall, url.Values{
				"idsite":      []string{h.piwikId},
				"action_name": []string{"Catgi/" + r.URL.String()},
				"apiv":        []string{"1"},
				"rec":         []string{"1"},
				"url":         []string{r.Host + r.URL.String()},
				"ref":         []string{r.Referer()},
				"rand":        []string{snowflake},
				"ua":          []string{r.UserAgent()},
				"lang":        []string{r.Header.Get("Accept-Language")},
				"cs":          []string{"utf8"},
				"urlref":      []string{r.Referer()},
			})
			if err != nil {
				log.Error(err)
			} else {
				log.Debug("Piwik responded with ", resp.StatusCode)
			}
		}()

		log.Debug("Calling next middlwere")
	}

	h.next.ServeHTTP(w, r)
}
