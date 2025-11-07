package auditapp

import (
	"net/http"

	"github.com/h2a26/go-firstcup/app/sdk/auth"
	"github.com/h2a26/go-firstcup/app/sdk/authclient"
	"github.com/h2a26/go-firstcup/app/sdk/mid"
	"github.com/h2a26/go-firstcup/business/domain/auditbus"
	"github.com/h2a26/go-firstcup/foundation/logger"
	"github.com/h2a26/go-firstcup/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	AuditBus   *auditbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newApp(cfg.AuditBus)

	app.HandlerFunc(http.MethodGet, version, "/audits", api.query, authen, ruleAdmin)
}
