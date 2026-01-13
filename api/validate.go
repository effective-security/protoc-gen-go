package api

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/xlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/protoc-gen-go/api", "api")

type Validator interface {
	Validate(ctx context.Context) error
}

// ValidateRequest validates a protobuf request message using its MessageDescription metadata.
// It checks for required fields, string length constraints, and numeric value constraints.
// The fields in the MessageDescription are annotated with the validation constraints:
/*
	// Required is the option for the field to be required.
	Required bool
	// RequiredOr is the option for the field to be required, if one of the
	// other values is provided.
	RequiredOr []string
	// Min is the option for the field minimum length for strings, and minimum
	// value for numbers.
	Min int32
	// Max is the option for the field maximum length for strings, and maximum
	// value for numbers.
	Max int32
	// MinCount is the option for the field minimum count for lists.
	MinCount int32
	// MaxCount is the option for the field maximum count for lists.
	MaxCount int32
*/

func ValidateRequest(ctx context.Context, req proto.Message, md *MessageDescription) (err error) {
	if req == nil {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: request cannot be nil", md.Name)
	}

	if md == nil {
		// If no metadata is provided, skip validation
		return nil
	}

	msgReflect := req.ProtoReflect()
	if !msgReflect.IsValid() {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: is not a valid protobuf message", md.GetDisplayName())
	}

	defer func() {
		if r := recover(); r != nil {
			logger.ContextKV(ctx, xlog.ERROR,
				"reason", "panic",
				"struct", md.Name,
				"err", r,
				"stack", string(debug.Stack()))
			// in case of panic, we set err to nil to avoid returning the panic error
			err = nil
		}
	}()

	err = validateReflectFields(ctx, msgReflect, md.Fields, "")
	if err != nil {
		logger.ContextKV(ctx, xlog.WARNING,
			"reason", "validateReflectFields",
			"struct", md.Name,
			"err", err,
		)
	}
	return
}

func validateReflectFields(ctx context.Context, msgReflect protoreflect.Message, fields []*FieldMeta, prefix string) error {
	pfields := msgReflect.Descriptor().Fields()
	for _, field := range fields {
		fd := pfields.ByName(protoreflect.Name(field.Name))
		if fd == nil {
			continue
		}

		v := msgReflect.Get(fd)
		fieldPath := field.Name
		if prefix != "" {
			fieldPath = prefix + "." + field.Name
		}

		valuePresent := hasFieldValue(msgReflect, fd, v)

		// Check RequiredOr fields
		if len(field.RequiredOr) > 0 {
			if valuePresent {
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
				if hasFieldValue(msgReflect, reqOrFd, reqOrValue) {
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
			if !valuePresent {
				return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s is required", fieldPath)
			}
		}

		if fd.IsList() {
			if err := validateListField(ctx, v, fd, field, fieldPath); err != nil {
				return err
			}
			continue
		}

		if fd.IsMap() {
			if err := validateMapField(ctx, v, fd, field, fieldPath); err != nil {
				return err
			}
			continue
		}

		if err := validateSingularValue(ctx, v, fd.Kind(), field, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

func validateListField(ctx context.Context, listValue protoreflect.Value, fd protoreflect.FieldDescriptor, field *FieldMeta, fieldPath string) error {
	plist := listValue.List()
	length := plist.Len()

	if field.MinCount > 0 && length < int(field.MinCount) {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum count is %d", fieldPath, field.MinCount)
	}
	if field.MaxCount > 0 && length > int(field.MaxCount) {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum count is %d", fieldPath, field.MaxCount)
	}

	if fd.Kind() == protoreflect.MessageKind && len(field.Fields) > 0 {
		for i := 0; i < length; i++ {
			elementPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			if err := validateReflectFields(ctx, plist.Get(i).Message(), field.Fields, elementPath); err != nil {
				return err
			}
		}
		return nil
	}

	for i := 0; i < length; i++ {
		elementPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := validateSingularValue(ctx, plist.Get(i), fd.Kind(), field, elementPath); err != nil {
			return err
		}
	}
	return nil
}

func validateMapField(ctx context.Context, mapValue protoreflect.Value, fd protoreflect.FieldDescriptor, field *FieldMeta, fieldPath string) error {
	pmap := mapValue.Map()
	length := pmap.Len()

	if field.MinCount > 0 && length < int(field.MinCount) {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum count is %d", fieldPath, field.MinCount)
	}
	if field.MaxCount > 0 && length > int(field.MaxCount) {
		return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum count is %d", fieldPath, field.MaxCount)
	}

	mv := fd.MapValue()
	var err error
	pmap.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		elementPath := fmt.Sprintf("%s[%s]", fieldPath, key.String())
		if mv.Kind() == protoreflect.MessageKind && len(field.Fields) > 0 {
			err = validateReflectFields(ctx, val.Message(), field.Fields, elementPath)
		} else {
			err = validateSingularValue(ctx, val, mv.Kind(), field, elementPath)
		}
		return err == nil
	})
	if err != nil {
		return err
	}
	return nil
}

func validateSingularValue(ctx context.Context, fieldValue protoreflect.Value, kind protoreflect.Kind, field *FieldMeta, fieldPath string) error {
	if kind == protoreflect.StringKind {
		strVal := fieldValue.String()
		if field.Min > 0 && len(strVal) < int(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum length is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && len(strVal) > int(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum length is %d", fieldPath, field.Max)
		}
	}

	if kind == protoreflect.BytesKind {
		bytesVal := fieldValue.Bytes()
		if field.Min > 0 && len(bytesVal) < int(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum length is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && len(bytesVal) > int(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum length is %d", fieldPath, field.Max)
		}
	}

	if err := checkNumericConstraints(ctx, fieldValue, kind, field, fieldPath); err != nil {
		return err
	}

	if kind == protoreflect.MessageKind && len(field.Fields) > 0 {
		msgVal := fieldValue.Message()
		if msgVal.IsValid() {
			if err := validateReflectFields(ctx, msgVal, field.Fields, fieldPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func hasFieldValue(msg protoreflect.Message, fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
	if msg.Has(fd) {
		return true
	}
	if !value.IsValid() {
		return false
	}
	return !isEmpty(value, fd)
}

// Helper function to check if a value is considered empty
func isEmpty(value protoreflect.Value, fd protoreflect.FieldDescriptor) bool {
	if !value.IsValid() {
		return true
	}
	if fd.IsList() {
		return value.List().Len() == 0
	}
	if fd.IsMap() {
		return value.Map().Len() == 0
	}

	switch fd.Kind() {
	case protoreflect.StringKind:
		return value.String() == ""
	case protoreflect.BytesKind:
		return len(value.Bytes()) == 0
	case protoreflect.MessageKind:
		return !value.Message().IsValid()
	// case protoreflect.BoolKind:
	// 	return !value.Bool()
	// case protoreflect.EnumKind:
	// 	return value.Enum() == 0
	default:
		return false
	}
}

func checkNumericConstraints(ctx context.Context, fieldValue protoreflect.Value, kind protoreflect.Kind, field *FieldMeta, fieldPath string) error {
	switch kind {
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		val := fieldValue.Int()
		if field.Min != 0 && val < int64(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum value is %d", fieldPath, field.Min)
		}
		if field.Max != 0 && val > int64(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum value is %d", fieldPath, field.Max)
		}
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		val := fieldValue.Uint()
		if field.Min > 0 && val < uint64(field.Min) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: minimum value is %d", fieldPath, field.Min)
		}
		if field.Max > 0 && val > uint64(field.Max) {
			return httperror.NewGrpcFromCtx(ctx, codes.InvalidArgument, "%s: maximum value is %d", fieldPath, field.Max)
		}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
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
