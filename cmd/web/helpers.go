package main

import "net/http"

func (app *application) serverError(w http.ResponseWriter , r *http.Request , err error) {
	var (
		method = r.Method
		url = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(),"method", method , "url" , url)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

func (app *application) clientError(w http.ResponseWriter , r *http.Request , status int) {
	http.Error(w , http.StatusText(status), status)
}
