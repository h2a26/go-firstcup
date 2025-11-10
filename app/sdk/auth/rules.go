package auth

import (
	_ "embed"
)

// These are the current set of rules we have for auth.
const (
	RuleAuthenticate   = "auth"                  //Verifies the token to validate the sender's token is valid.
	RuleAny            = "rule_any"              //Allows any authenticated user (has at least one role)
	RuleAdminOnly      = "rule_admin_only"       //Only allows users with the "ADMIN" role
	RuleUserOnly       = "rule_user_only"        //Only allows users with the "USER" role
	RuleAdminOrSubject = "rule_admin_or_subject" //Allows either Users with the "ADMIN" role, OR Regular users accessing their own data (UserID matches Subject)
)

// Package name of our rego code.
const (
	opaPackage string = "ardan.rego"
)

// Core OPA policies.
var (
	//go:embed rego/authentication.rego
	regoAuthentication string

	//go:embed rego/authorization.rego
	regoAuthorization string
)
