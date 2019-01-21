package api

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

type ConnectRequest struct {
	Host string
}

type ConnectResponse struct {}

func (api *APIDispatcher) ConnectHandler() (func(ctx *fasthttp.RequestCtx)){
	return func(ctx *fasthttp.RequestCtx) {
		log.Debug().
			Str("Host", string(ctx.Host())).
			Msg("Connection request.")
		fmt.Fprint(ctx, "Here are your infos!\n")
	}
}