package web

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
)

// Respond marshals a value to JSON and sends it to the client.
func Respond(ctx context.Context, writer http.ResponseWriter, val interface{}, statusCode int) error {
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return errors.New("web values missing from context")
	}

	v.StatusCode = statusCode

	if statusCode == http.StatusNoContent {
		writer.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Convert the response value to JSON
	data, err := json.Marshal(val)
	if err != nil {
		return errors.Wrap(err, "marshalling value to json")
	}

	writer.Header().Set("content-type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)

	if _, err := writer.Write(data); err != nil {
		return errors.Wrap(err, "writing to client")
	}

	return nil
}

// RespondError knows how to handle errors going out to the client.
func RespondError(ctx context.Context, writer http.ResponseWriter, err error) error {
	// If the error was of the type *Error, the handler has
	// a specific status code and error to return
	if webErr, ok := errors.Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		if err := Respond(ctx, writer, er, webErr.Status); err != nil {
			return err
		}
		return nil
	}

	// If not, the handler sent any arbitrary error value so use 500
	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}
	if err := Respond(ctx, writer, er, http.StatusInternalServerError); err != nil {
		return err
	}
	return nil
}
