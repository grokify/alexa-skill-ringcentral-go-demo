package rcskillserver

import (
	"encoding/json"
	"fmt"

	"github.com/buaazp/fasthttprouter"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/grokify/alexa-skill-ringcentral-go/src/config"
	"github.com/grokify/alexa-skill-ringcentral-go/src/intents/ringout"
	"github.com/grokify/alexa-skill-ringcentral-go/src/intents/sms"
	"github.com/grokify/alexa-skill-ringcentral-go/src/intents/sms_body"
	"github.com/grokify/alexa-skill-ringcentral-go/src/intents/voicemail_count"
	alexa "github.com/mikeflynn/go-alexa/skillserver"
)

const (
	RouteRingCentral = "/echo/ringcentral"
)

type Handler struct {
	Configuration config.Configuration
}

func NewHandler(cfg config.Configuration) Handler {
	cfg.AddressBook.Inflate()
	return Handler{Configuration: cfg}
}

func (h *Handler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	echoReq := &alexa.EchoRequest{}
	json.Unmarshal(ctx.PostBody(), echoReq)

	log.WithFields(log.Fields{
		"sessionID": echoReq.Session.SessionID,
		"body":      fmt.Sprintf("%v", ctx.PostBody())}).Info("HandleFastHTTP Request")

	var echoResp *alexa.EchoResponse

	switch echoReq.GetIntentName() {
	case "RingCentralGetNewVoicemailCountIntent":
		echoResp = voicemailcount.HandleRequest(h.Configuration, echoReq)
	case "RingCentralSendSMSIntent":
		echoResp = sms.HandleRequest(h.Configuration, echoReq)
	case "RingCentralSendSMSIntentBody":
		echoResp = smsbody.HandleRequest(h.Configuration, echoReq)
	case "RingCentralCreateRingOutIntent":
		echoResp = ringout.HandleRequest(h.Configuration, echoReq)
	default:
		echoResp = alexa.NewEchoResponse().OutputSpeech("I'm sorry, I didn't get that. Can you say that again?").EndSession(false)
	}

	echoRespBytes, err := json.Marshal(echoResp)
	if err != nil {
		log.WithFields(log.Fields{
			"sessionID": echoReq.Session.SessionID,
			"body":      fmt.Sprintf("%v", ctx.PostBody())}).Warn("HandleFastHTTP Response JSON Marshal Error")
	}
	log.WithFields(log.Fields{
		"sessionID": echoReq.Session.SessionID,
		"body":      string(echoRespBytes)}).Info("HandleFastHTTP Response")

	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprintln(ctx, string(echoRespBytes))
}

// StartServer initializes and starts the webhook proxy server
func StartServer(cfg config.Configuration) {
	log.SetLevel(log.InfoLevel)

	router := fasthttprouter.New()

	rcHandler := NewHandler(cfg)
	router.POST(RouteRingCentral, rcHandler.HandleFastHTTP)

	log.WithFields(log.Fields{}).Info(fmt.Sprintf("Listening on port %v.", cfg.Port))
	log.Fatal(fasthttp.ListenAndServe(cfg.Address(), router.Handler))
}
