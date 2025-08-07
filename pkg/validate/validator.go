package validator

import (
	"context"
	validatorPkg "github.com/go-playground/validator/v10"
	"github.com/omniful/go_commons/jwt/public"
)

var validatePkg *validatorPkg.Validate

func Get() *validatorPkg.Validate {
	return validatePkg
}

func Set() {
	validatePkg = validatorPkg.New()
	err := validatePkg.RegisterValidationCtx("tenant_id", validateTenantIDFunc)
	if err != nil {
		return
	}
}

func validateTenantIDFunc(ctx context.Context, f1 validatorPkg.FieldLevel) bool {
	tenantID, err := public.GetTenantID(ctx)
	if err != nil {
		return false
	}

	return f1.Field().String() == tenantID
}
