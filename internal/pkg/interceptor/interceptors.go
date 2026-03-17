package interceptor

import (
	"context"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/Chaice1/Order-Service/internal/pkg/auth"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func ValidationInterceptor(v protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if message, ok := req.(proto.Message); ok {
			if err := v.Validate(message); err != nil {
				st := status.New(codes.InvalidArgument, codes.InvalidArgument.String())
				st, _ = st.WithDetails(&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{
							Field:       "request",
							Description: err.Error(),
						},
					},
				})
				return nil, st.Err()
			}
		}
		return handler(ctx, req)
	}
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if info.FullMethod == "/proto.go.user.UserService/Login" || info.FullMethod == "/proto.go.user.UserService/Register" {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md["authorization"]) == 0 {
			return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
		}
		tokenslice := strings.Split(md["authorization"][0], " ")
		if tokenslice[0] != "Bearer" || len(tokenslice) != 2 {
			return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
		}

		claims, err := auth.ValidateToken(tokenslice[1])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
		}

		newctx := context.WithValue(ctx, "user_id", claims.UserId)

		newcttx := context.WithValue(newctx, "role", claims.Role)

		return handler(newcttx, req)
	}
}
