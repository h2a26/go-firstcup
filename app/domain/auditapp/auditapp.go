// Package auditapp maintains the app layer api for the audit domain.
package auditapp

import (
	"context"
	"net/http"

	"github.com/h2a26/go-firstcup/app/sdk/errs"
	"github.com/h2a26/go-firstcup/app/sdk/query"
	"github.com/h2a26/go-firstcup/business/domain/auditbus"
	"github.com/h2a26/go-firstcup/business/domain/userbus"
	"github.com/h2a26/go-firstcup/business/sdk/order"
	"github.com/h2a26/go-firstcup/business/sdk/page"
	"github.com/h2a26/go-firstcup/foundation/web"
)

type app struct {
	auditBus *auditbus.Business
}

func newApp(auditBus *auditbus.Business) *app {
	return &app{
		auditBus: auditBus,
	}
}

func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return errs.NewFieldErrors("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return err.(*errs.Error)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, userbus.DefaultOrderBy)
	if err != nil {
		return errs.NewFieldErrors("order", err)
	}

	adts, err := a.auditBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.auditBus.Count(ctx, filter)
	if err != nil {
		return errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(toAppAudits(adts), total, page)
}
