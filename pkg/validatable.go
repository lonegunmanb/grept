package pkg

import (
	"io/fs"
	"strconv"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New(validator.WithRequiredStructEnabled())

func registerValidator() {
	_ = Validate.RegisterValidation("conflict_with", validateConflictWith)
	_ = Validate.RegisterValidation("at_least_one_of", validateAtLeastOneOf)
	_ = Validate.RegisterValidation("required_with", validateRequiredWith)
	_ = Validate.RegisterValidation("all_string_in_slice", validateAllStringInSlice)
	_ = Validate.RegisterValidation("file_mode", validateFileMode)
}

func validateFileMode(fl validator.FieldLevel) bool {
	field := fl.Field().Interface()
	num, ok := field.(fs.FileMode)
	if !ok {
		return false
	}
	mode, err := strconv.ParseUint(strconv.Itoa(int(num)), 8, 32)
	if err != nil {
		return false
	}
	return uint32(mode) <= uint32(fs.ModePerm)
}

func validateAllStringInSlice(fl validator.FieldLevel) bool {
	candidatesQuery := linq.From(strings.Split(fl.Param(), " "))
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	values, ok := thisField.Interface().([]string)
	if !ok {
		return false
	}
	valuesQuery := linq.From(values)

	return !valuesQuery.Except(candidatesQuery).Any()
}

func validateConflictWith(fl validator.FieldLevel) bool {
	conflictWith := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	for _, anotherField := range conflictWith {
		field := parentStruct.FieldByName(anotherField)
		if field.IsValid() && !field.IsZero() {
			return false
		}
	}
	return true
}

func validateRequiredWith(fl validator.FieldLevel) bool {
	requiredWith := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	fieldName := fl.FieldName()
	thisField := parentStruct.FieldByName(fieldName)
	if !thisField.IsValid() || thisField.IsZero() {
		return true
	}

	for _, anotherField := range requiredWith {
		field := parentStruct.FieldByName(anotherField)
		if !field.IsValid() || field.IsZero() {
			return false
		}
	}
	return true
}

func validateAtLeastOneOf(fl validator.FieldLevel) bool {
	atLeastOneOf := strings.Split(fl.Param(), " ")
	parentStruct := fl.Parent()
	for _, fieldName := range atLeastOneOf {
		field := parentStruct.FieldByName(fieldName)
		if field.IsValid() && !field.IsZero() {
			return true
		}
	}
	return false
}
