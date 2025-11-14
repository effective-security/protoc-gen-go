package api

import (
	"context"
	"strings"

	"github.com/effective-security/porto/xhttp/httperror"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Validator interface {
	Validate(ctx context.Context) error
}

// ValidateRequest validates a protobuf request message using its MessageDescription metadata.
// It checks for required fields, string length constraints, and numeric value constraints.
func ValidateRequest(ctx context.Context, req proto.Message, md *MessageDescription) error {
	if req == nil {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: request cannot be nil", md.Name)
	}

	if md == nil {
		// If no metadata is provided, skip validation
		return nil
	}

	msgReflect := req.ProtoReflect()
	if !msgReflect.IsValid() {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: is not a valid protobuf message", md.Display)
	}

	return validateReflectFields(ctx, msgReflect, md.Fields, "")
}

func validateReflectFields(ctx context.Context, msgReflect protoreflect.Message, fields []*FieldMeta, prefix string) error {
	pfields := msgReflect.Descriptor().Fields()
	for _, field := range fields {
		fd := pfields.ByName(protoreflect.Name(field.Name))
		if fd == nil {
			continue
		}
		kind := fd.Kind()

		v := msgReflect.Get(fd)
		fieldPath := field.Name
		if prefix != "" {
			fieldPath = prefix + "." + field.Name
		}

		// Check RequiredOr fields
		if len(field.RequiredOr) > 0 {
			if v.IsValid() && !isEmpty(v, fd) {
				continue
			}
			// Check if any of the RequiredOr fields are valid
			isValid := false
			for _, reqOrField := range field.RequiredOr {
				reqOrFd := pfields.ByName(protoreflect.Name(reqOrField))
				if reqOrFd == nil {
					continue
				}
				reqOrValue := msgReflect.Get(reqOrFd)
				if reqOrValue.IsValid() && !isEmpty(reqOrValue, reqOrFd) {
					isValid = true
					break
				}
			}
			if !isValid {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: at least one of the fields must be set: %s", fieldPath, strings.Join(field.RequiredOr, ", "))
			}
		}

		// Check required fields
		if field.Required {
			if err := checkProtoRequired(ctx, v, fd, fieldPath); err != nil {
				return err
			}
		}

		// Check string length constraints
		if kind == protoreflect.StringKind {
			strVal := v.String()
			if field.Min > 0 && len(strVal) < int(field.Min) {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum length is %d", fieldPath, field.Min)
			}
			if field.Max > 0 && len(strVal) > int(field.Max) {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum length is %d", fieldPath, field.Max)
			}
		}

		// Check numeric constraints
		if err := checkNumericConstraints(ctx, v, kind, field, fieldPath); err != nil {
			return err
		}

		// Check array/slice constraints
		if fd.IsList() {
			plist := v.List()
			if field.MinCount > 0 && plist.Len() < int(field.MinCount) {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum count is %d", fieldPath, field.MinCount)
			}
			if field.MaxCount > 0 && plist.Len() > int(field.MaxCount) {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum count is %d", fieldPath, field.MaxCount)
			}
		}

		// Recursively validate nested structs
		if kind == protoreflect.MessageKind && len(field.Fields) > 0 {
			if err := validateReflectFields(ctx, v.Message(), field.Fields, fieldPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function to check if a value is considered empty
func isEmpty(value protoreflect.Value, fd protoreflect.FieldDescriptor) bool {
	kind := fd.Kind()
	switch kind {
	case protoreflect.StringKind:
		return value.String() == ""
	// case protoreflect.RepeatedKind:
	// 	return value.List().Len() == 0
	case protoreflect.MessageKind:
		return !value.Message().IsValid()
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.FloatKind:
		// Typically, zero values might be considered empty, but depends on the context, here null or default is treated as empty
		return false // assuming numeric zero is not empty unless specified otherwise
	default:
		if fd.IsList() {
			return value.List().Len() == 0
		}
		if fd.IsMap() {
			return value.Map().Len() == 0
		}
		return false
	}
}

func checkProtoRequired(ctx context.Context, fieldValue protoreflect.Value, fieldDescriptor protoreflect.FieldDescriptor, fieldPath string) error {
	switch fieldDescriptor.Kind() {
	case protoreflect.StringKind:
		if fieldValue.String() == "" {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldPath)
		}
	case protoreflect.BytesKind:
		if fieldValue.Bytes() == nil || len(fieldValue.Bytes()) == 0 {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldPath)
		}
	case protoreflect.MessageKind:
		if !fieldValue.Message().IsValid() {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldPath)
		}
	// case protoreflect.ListKind:
	// 	if fieldValue.Len() == 0 {
	// 		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldPath)
	// 	}
	case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.FloatKind:
		// For numeric fields, we don't check for zero as it might be a valid value
		// unless explicitly configured with min/max constraints
	default:
		// For other types, we don't enforce required check by default
	}
	return nil
}

func checkNumericConstraints(ctx context.Context, fieldValue protoreflect.Value, kind protoreflect.Kind, field *FieldMeta, fieldPath string) error {
	switch kind {
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		val := fieldValue.Int()
		if field.Min > 0 && val < int64(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum value is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && val > int64(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum value is %d", fieldPath, field.Max)
		}
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		val := fieldValue.Uint()
		if field.Min > 0 && val < uint64(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum value is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && val > uint64(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum value is %d", fieldPath, field.Max)
		}
	case protoreflect.FloatKind:
		val := fieldValue.Float()
		if field.Min > 0 && val < float64(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum value is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && val > float64(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum value is %d", fieldPath, field.Max)
		}
	}
	return nil
}

// // SanitizeString removes potentially dangerous characters and limits string length.
// // This provides an additional layer of defense against injection attacks.
// func SanitizeString(s string, maxLen int) string {
// 	// Remove null bytes
// 	s = strings.ReplaceAll(s, "\x00", "")

// 	// Trim whitespace
// 	s = strings.TrimSpace(s)

// 	// Limit length
// 	if maxLen > 0 && len(s) > maxLen {
// 		s = s[:maxLen]
// 	}

// 	return s
// }

// // ValidateBytes checks if byte array meets basic requirements.
// func ValidateBytes(ctx context.Context, data []byte, fieldName string, maxSize int) error {
// 	if len(data) == 0 {
// 		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldName)
// 	}

// 	if maxSize > 0 && len(data) > maxSize {
// 		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s exceeds maximum size of %d bytes", fieldName, maxSize)
// 	}

// 	return nil
// }

// // FormatValidationError formats a validation error with context.
// func FormatValidationError(ctx context.Context, field string, msg string) error {
// 	return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "validation failed for %s: %s", field, msg)
// }
