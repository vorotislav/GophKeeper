package responder

import "net/http"

type httpStatusCoder interface {
	StatusCode() int
}

type jsoner interface {
	ToJSON() ([]byte, error)
}

// JSON writes passed model as a json-formatted string into writer.
func JSON(w http.ResponseWriter, model jsoner) {
	if name := "Content-Type"; w.Header().Get(name) == "" {
		w.Header().Set(name, "application/json; charset=utf-8")
	}

	content, err := model.ToJSON()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`"service response cannot be converted into json format"`)) // fallback

		return
	}

	var code = http.StatusOK // default code

	if v, ok := model.(httpStatusCoder); ok {
		if statusCode := v.StatusCode(); statusCode != 0 {
			code = statusCode // override with model code
		}
	}

	w.WriteHeader(code)

	_, _ = w.Write(content)
}
