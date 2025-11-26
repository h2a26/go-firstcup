package vproductbus

import (
	"time"

	"github.com/google/uuid"

	"github.com/h2a26/go-firstcup/business/types/money"
	"github.com/h2a26/go-firstcup/business/types/name"
	"github.com/h2a26/go-firstcup/business/types/quantity"
)

// Product represents an individual product with extended information.
type Product struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        name.Name
	Cost        money.Money
	Quantity    quantity.Quantity
	DateCreated time.Time
	DateUpdated time.Time
	UserName    name.Name
}
