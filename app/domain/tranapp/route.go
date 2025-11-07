package tranapp

import (
	"net/http"

	"github.com/h2a26/go-firstcup/app/sdk/auth"
	"github.com/h2a26/go-firstcup/app/sdk/authclient"
	"github.com/h2a26/go-firstcup/app/sdk/mid"
	"github.com/h2a26/go-firstcup/business/domain/productbus"
	"github.com/h2a26/go-firstcup/business/domain/userbus"
	"github.com/h2a26/go-firstcup/business/sdk/sqldb"
	"github.com/h2a26/go-firstcup/foundation/logger"
	"github.com/h2a26/go-firstcup/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log        *logger.Logger
	DB         *sqlx.DB
	UserBus    userbus.ExtBusiness
	ProductBus *productbus.Business
	AuthClient *authclient.Client
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	transaction := mid.BeginCommitRollback(cfg.Log, sqldb.NewBeginner(cfg.DB))
	ruleAdmin := mid.Authorize(cfg.AuthClient, auth.RuleAdminOnly)

	api := newApp(cfg.UserBus, cfg.ProductBus)

	app.HandlerFunc(http.MethodPost, version, "/tranexample", api.create, authen, ruleAdmin, transaction)
}
