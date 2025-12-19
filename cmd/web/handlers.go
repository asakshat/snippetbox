package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/asakshat/snippetbox/internal/models"
	"github.com/asakshat/snippetbox/internal/validator"
	"github.com/asakshat/snippetbox/ui/components/pages"
)

func isHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	data := app.newTemplateData(r)
	component := pages.Home(data, snippets)
	component.Render(r.Context(), w)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	component := pages.ViewSnippet(data, snippet)
	component.Render(r.Context(), w)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	form := pages.SnippetCreateForm{
		Expires: 365,
	}
	component := pages.CreateSnippet(data, form)
	component.Render(r.Context(), w)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form pages.SnippetCreateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validator.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.Validator.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters")
	form.Validator.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.Validator.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Validator.Valid() {
		if isHtmxRequest(r) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			component := pages.CreateForm(app.getCSRFToken(r), form)
			component.Render(r.Context(), w)
			return
		}
		
		data := app.newTemplateData(r)
		component := pages.CreateSnippet(data, form)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component.Render(r.Context(), w)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	
	
	if isHtmxRequest(r) {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/snippet/view/%d", id))
		w.WriteHeader(http.StatusOK)
		return
	}
	
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	form := pages.UserLoginForm{}
	component := pages.Login(data, form)
	component.Render(r.Context(), w)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form pages.UserLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validator.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.Validator.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.Validator.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Validator.Valid() {
		if isHtmxRequest(r) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			component := pages.LoginForm(app.getCSRFToken(r), form)
			component.Render(r.Context(), w)
			return
		}
		
		data := app.newTemplateData(r)
		component := pages.Login(data, form)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component.Render(r.Context(), w)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Validator.AddNonFieldError("Email or password is incorrect")
			
			if isHtmxRequest(r) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				component := pages.LoginForm(app.getCSRFToken(r), form)
				component.Render(r.Context(), w)
				return
			}
			
			data := app.newTemplateData(r)
			component := pages.Login(data, form)
			w.WriteHeader(http.StatusUnprocessableEntity)
			component.Render(r.Context(), w)
			return
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	
	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path == "" {
		path = "/snippet/create"
	}
	
	if isHtmxRequest(r) {
		w.Header().Set("HX-Redirect", path)
		w.WriteHeader(http.StatusOK)
		return
	}
	
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	
	if isHtmxRequest(r) {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) aboutUs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	component := pages.About(data)
	component.Render(r.Context(), w)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	data := app.newTemplateData(r)
	component := pages.Account(data, user)
	component.Render(r.Context(), w)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	form := pages.AccountPasswordUpdateForm{}
	component := pages.PasswordUpdate(data, form)
	component.Render(r.Context(), w)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form pages.AccountPasswordUpdateForm
	
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validator.CheckField(validator.NotBlank(form.CurrentPassword), "current_password", "This field cannot be blank")
	form.Validator.CheckField(validator.NotBlank(form.NewPassword), "new_password", "This field cannot be blank")
	form.Validator.CheckField(validator.NotBlank(form.ConfirmNewPassword), "confirm_password", "This field cannot be blank")
	form.Validator.CheckField(validator.MinChars(form.NewPassword, 8), "new_password", "This field must be at least 8 characters long")
	form.Validator.CheckField(form.NewPassword == form.ConfirmNewPassword, "new_password", "Passwords do not match")

	if !form.Validator.Valid() {
		if isHtmxRequest(r) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			component := pages.PasswordUpdateForm(app.getCSRFToken(r), form)
			component.Render(r.Context(), w)
			return
		}
		
		data := app.newTemplateData(r)
		component := pages.PasswordUpdate(data, form)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component.Render(r.Context(), w)
		return
	}

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	err = app.users.UpdatePassword(userID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Validator.AddFieldError("current_password", "Current password is incorrect")
			
			if isHtmxRequest(r) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				component := pages.PasswordUpdateForm(app.getCSRFToken(r), form)
				component.Render(r.Context(), w)
				return
			}
			
			data := app.newTemplateData(r)
			component := pages.PasswordUpdate(data, form)
			w.WriteHeader(http.StatusUnprocessableEntity)
			component.Render(r.Context(), w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated!")
	
	if isHtmxRequest(r) {
		w.Header().Set("HX-Redirect", "/account/view")
		w.WriteHeader(http.StatusOK)
		return
	}
	
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	form := pages.UserSignupForm{}
	component := pages.Signup(data, form)
	component.Render(r.Context(), w)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form pages.UserSignupForm
	
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.Validator.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.Validator.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.Validator.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.Validator.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.Validator.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Validator.Valid() {
		data := app.newTemplateData(r)
		component := pages.Signup(data, form)
		w.WriteHeader(http.StatusUnprocessableEntity)
		component.Render(r.Context(), w)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Validator.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			component := pages.Signup(data, form)
			w.WriteHeader(http.StatusUnprocessableEntity)
			component.Render(r.Context(), w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
