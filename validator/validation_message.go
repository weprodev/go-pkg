package validator

import "github.com/go-playground/validator/v10"

func FieldErrorToString(fe validator.FieldError) string {
	switch fe.Tag() {
	// Basic validations
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email address"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must not exceed " + fe.Param() + " characters"
	case "len":
		return fe.Field() + " must be exactly " + fe.Param() + " characters"

	// String validations
	case "alpha":
		return fe.Field() + " must contain only letters"
	case "alphanum":
		return fe.Field() + " must contain only letters and numbers"
	case "alphaunicode":
		return fe.Field() + " must contain only unicode letters"
	case "alphanumunicode":
		return fe.Field() + " must contain only unicode letters and numbers"
	case "ascii":
		return fe.Field() + " must contain only ascii characters"
	case "contains":
		return fe.Field() + " must contain the text '" + fe.Param() + "'"
	case "containsany":
		return fe.Field() + " must contain at least one of the characters '" + fe.Param() + "'"
	case "containsrune":
		return fe.Field() + " must contain the rune '" + fe.Param() + "'"
	case "excludes":
		return fe.Field() + " must not contain the text '" + fe.Param() + "'"
	case "excludesall":
		return fe.Field() + " must not contain any of the characters '" + fe.Param() + "'"
	case "excludesrune":
		return fe.Field() + " must not contain the rune '" + fe.Param() + "'"
	case "lowercase":
		return fe.Field() + " must be lowercase"
	case "uppercase":
		return fe.Field() + " must be uppercase"
	case "startswith":
		return fe.Field() + " must start with '" + fe.Param() + "'"
	case "endswith":
		return fe.Field() + " must end with '" + fe.Param() + "'"

	// Number validations
	case "gt":
		return fe.Field() + " must be greater than " + fe.Param()
	case "gte":
		return fe.Field() + " must be greater than or equal to " + fe.Param()
	case "lt":
		return fe.Field() + " must be less than " + fe.Param()
	case "lte":
		return fe.Field() + " must be less than or equal to " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of [" + fe.Param() + "]"

	// Date validations
	case "datetime":
		return fe.Field() + " must be a valid datetime"
	case "date":
		return fe.Field() + " must be a valid date"

	// Network validations
	case "ip":
		return fe.Field() + " must be a valid IP address"
	case "ipv4":
		return fe.Field() + " must be a valid IPv4 address"
	case "ipv6":
		return fe.Field() + " must be a valid IPv6 address"
	case "tcp_addr":
		return fe.Field() + " must be a valid TCP address"
	case "tcp4_addr":
		return fe.Field() + " must be a valid TCP4 address"
	case "tcp6_addr":
		return fe.Field() + " must be a valid TCP6 address"
	case "udp_addr":
		return fe.Field() + " must be a valid UDP address"
	case "udp4_addr":
		return fe.Field() + " must be a valid UDP4 address"
	case "udp6_addr":
		return fe.Field() + " must be a valid UDP6 address"
	case "unix_addr":
		return fe.Field() + " must be a valid Unix address"

	// Format validations
	case "uuid":
		return fe.Field() + " must be a valid UUID"
	case "uuid3":
		return fe.Field() + " must be a valid UUID v3"
	case "uuid4":
		return fe.Field() + " must be a valid UUID v4"
	case "uuid5":
		return fe.Field() + " must be a valid UUID v5"
	case "json":
		return fe.Field() + " must be a valid JSON string"
	case "base64":
		return fe.Field() + " must be a valid base64 string"
	case "base64url":
		return fe.Field() + " must be a valid base64url string"
	case "jwt":
		return fe.Field() + " must be a valid JWT token"

	// Boolean validations
	case "boolean":
		return fe.Field() + " must be a valid boolean value"

	// Custom validations for our domain
	case "gender":
		return fe.Field() + " must be one of: male, female, other, unknown"
	case "user_status":
		return fe.Field() + " must be one of: pending, accepted, blocked, suspended"
	case "service_name":
		return fe.Field() + " must be a valid service name"

	default:
		return fe.Field() + " is invalid"
	}
}
