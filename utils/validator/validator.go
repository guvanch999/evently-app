package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/nyaruka/phonenumbers"
	"log"
	"net/http"
)

type StructValidator struct {
	validator *validator.Validate
}

func (cv *StructValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewValidator() *StructValidator {
	val := validator.New()
	err := val.RegisterValidation("phone_number", phoneNumberValidate)
	if err != nil {
		log.Fatal(err)
	}
	return &StructValidator{
		validator: val,
	}
}

func phoneNumberValidate(fl validator.FieldLevel) bool {
	_, err := phonenumbers.Parse(fl.Field().String(), "")
	if err != nil {
		return false
	}
	return true
}
