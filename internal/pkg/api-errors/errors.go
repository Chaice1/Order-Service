package apierrors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIerror struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var GRPCtohttp = map[codes.Code]struct {
	Httpstatus int
	Apicode    string
}{
	codes.AlreadyExists:    {http.StatusConflict, "ALREADY_EXISTS"},
	codes.PermissionDenied: {http.StatusForbidden, "FORBIDDEN"},
	codes.Internal:         {http.StatusInternalServerError, "INTERNAL"},
	codes.InvalidArgument:  {http.StatusBadRequest, "BAD_REQUEST"},
	codes.NotFound:         {http.StatusNotFound, "NOT_FOUND"},
	codes.Unauthenticated:  {http.StatusUnauthorized, "UNAUTHORIZED"},
}

func HandleErrors(c *gin.Context, err error) {

	st, ok := status.FromError(err)

	if ok {

		val, ok := GRPCtohttp[st.Code()]
		if !ok {

			val = struct {
				Httpstatus int
				Apicode    string
			}{Httpstatus: http.StatusInternalServerError, Apicode: "INTERNAL"}
		}

		c.JSON(val.Httpstatus, APIerror{
			Code:    val.Apicode,
			Message: st.Message(),
		})
		return
	}

	c.JSON(http.StatusBadRequest, APIerror{
		Code:    "BAD_REQUEST",
		Message: "request error",
	})

}
