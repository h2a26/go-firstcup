package auditbus

import "github.com/h2a26/go-firstcup/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByObjID, order.ASC)

// Set of fields that the results can be ordered by.
const (
	OrderByObjID     = "a"
	OrderByObjDomain = "b"
	OrderByObjName   = "c"
	OrderByActorID   = "d"
	OrderByAction    = "e"
)
