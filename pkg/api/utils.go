package api

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

func getIdentity(ctx *fasthttp.RequestCtx) string {
	return ctx.UserValue("identity").(string)
}

func getDomain(ctx *fasthttp.RequestCtx) string {
	return ctx.UserValue("domain").(string)
}

func respond(ctx *fasthttp.RequestCtx, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error().
			Str("error", err.Error()).
			Msg("Failed to marshal response")
		ctx.SetStatusCode(500)
		return err
	}

	if _, err := ctx.Write(b); err != nil {
		log.Warn().
			Str("error", err.Error()).
			Str("response", string(b)).
			Msg("Could not respond to client")
		return err
	}
	return nil
}

func failRequest(ctx *fasthttp.RequestCtx, msg string, code int) {
	ctx.SetStatusCode(code)
	resp := RegisterResponse{
		Err: &registerError{
			Msg:  msg,
			Code: code,
		},
	}
	respond(ctx, resp)
}

// ValidateDomain returns an error if the parameter is not a valid subdomain
// We only allow alphanumeric characters
func ValidateDomain(domain string) error {
	if match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", domain); !match {
		return errors.New("illegal characters in requested domain")
	}
	return nil
}
