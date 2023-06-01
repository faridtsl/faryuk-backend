package utils

import (
	"FaRyuk/internal/types"
	"encoding/json"
	"net/http"
)

// FIXME error messages/code should be clear and respect the HTTP standards
func writeResponse(w *http.ResponseWriter, m types.JSONReturn) {
	(*w).Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(*w).Encode(m)
	if err != nil {
		return
	}
}

func WriteForbidden(w *http.ResponseWriter, m string) {
	(*w).WriteHeader(http.StatusForbidden)
	writeResponse(w, types.JSONReturn{Status: "Forbidden", Body: m, Code: http.StatusForbidden})
}

func WriteUnAuthorized(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNotFound)
	writeResponse(w, types.JSONReturn{Status: "UnAuthorized", Code: http.StatusUnauthorized})
}

func WriteBadRequest(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNotFound)
	writeResponse(w, types.JSONReturn{Status: "BadRequest", Code: http.StatusBadRequest})
}

func WriteNotFound(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNotFound)
	writeResponse(w, types.JSONReturn{Status: "NotFound", Code: http.StatusNotFound})
}

func WriteInternalError(w *http.ResponseWriter, m string) {
	(*w).WriteHeader(http.StatusInternalServerError)
	writeResponse(w, types.JSONReturn{Status: "Fail", Body: m, Code: http.StatusInternalServerError})
}

func ReturnSuccess(w *http.ResponseWriter, m interface{}) {
	writeResponse(w, types.JSONReturn{Status: "Success", Code: http.StatusOK, Body: m})
}

func ReturnSuccessNoContent(w *http.ResponseWriter) {
	writeResponse(w, types.JSONReturn{Status: "Success", Code: http.StatusNoContent})
}

func ReturnSuccessCreated(w *http.ResponseWriter) {
	writeResponse(w, types.JSONReturn{Status: "Success", Code: http.StatusCreated})
}
