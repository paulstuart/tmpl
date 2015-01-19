package main

import (
	//"encoding/json"
	"fmt"
	//"net"
	"net/http"
	//"os"
	"strconv"
	"strings"
	"time"

	dbu "github.com/paulstuart/dbutil"
)

var (
	UCookie = "userinfo"
)

type Common struct {
	Title string
	User  *User
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Common
		Title string
		User  *User
	}{
		Title: "Users",
		User:  currentUser(r),
	}
	renderTemplate(w, r, "index", true, data)
}

func usersListPage(w http.ResponseWriter, r *http.Request) {
	Users := userList()
	data := struct {
		Common
		Title string
		Users []User
		User  *User
	}{
		Title: "Users",
		Users: Users,
		User:  currentUser(r),
	}
	renderTemplate(w, r, "user_list", true, data)
}

func UserEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var u User
		objFromForm(&u, r.Form)
		//	var action string
		if u.ID == 0 {
			if err := u.Add(); err != nil {
				fmt.Println("Add error", err)
			}
			//	action = "added"
		} else {
			//	action = "modified"
			u.Update()
		}
		//user := currentUser(r)
		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	} else {
		const pref = len(pathPrefix + "/user/edit/")
		var edit *User
		title := "Add User"
		if len(r.URL.Path) > pref {
			id := r.URL.Path[pref:]
			edit, _ = UserByID(id)
			title = "Edit User"
		}
		data := struct {
			Title    string
			User     *User
			EditUser *User
			//Levels   []userLevel
		}{
			Title:    title,
			User:     currentUser(r),
			EditUser: edit,
			//Levels:   userLevels,
		}
		renderTemplate(w, r, "user_edit", true, data)
	}
}

func pingPage(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	uptime := time.Since(start_time)
	stats := strings.Join(db.Stats(), "\n")
	fmt.Fprintf(w, "status: %s\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\ndb stats:\n%s\n", status, version, Hostname, start_time, uptime, stats)
}

func DebugPage(w http.ResponseWriter, r *http.Request) {
	const pref = len(pathPrefix + "/db/debug/")
	what := r.URL.Path[pref:]
	on, _ := strconv.ParseBool(what)
	fmt.Println("DEBUG?", what, "ON:", on)
	db.Debug = true
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "db debug: %t\n", on)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	/*
		c := &http.Cookie{Name: ACookie, Value: "", Path: "/", Expires: time.Unix(0, 0)}
		http.SetCookie(w, c)
	*/
	c := &http.Cookie{Name: UCookie, Value: "", Path: "/", Expires: time.Unix(0, 0)}
	http.SetCookie(w, c)
	http.Redirect(w, r, pathPrefix+"/", 302)
}

type Profile struct {
	Common
	Prefix, Prompt, ErrorMsg string
}

func loginPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	///data := struct{ Prefix, Prompt, ErrorMsg string }{Prefix: pathPrefix, Prompt: loginPrompt, ErrorMsg: errMsg}
	data := Profile{Prefix: pathPrefix, Prompt: loginPrompt, ErrorMsg: errMsg}
	renderTemplate(w, r, "login", false, data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		_, err := UserByEmail(username)
		if err != nil {
			loginPage(w, r, username+" is not authorized for access")
			return
		}

		if user := Authenticate(username, password); user != nil {
			fmt.Println("OK", username, password)
			expires := time.Now().Add(time.Minute * sessionMinutes)
			/*
				cookie := &http.Cookie{
					Name:    ACookie,
					Value:   ASecret,
					Path:    "/",
					Expires: expires,
				}
				http.SetCookie(w, cookie)
			*/

			u := &http.Cookie{
				Name:    uCookie,
				Path:    "/",
				Value:   user.ToCookie(),
				Expires: expires,
			}
			http.SetCookie(w, u)

			fmt.Println("REDIRECT", pathPrefix+"/")
			http.Redirect(w, r, pathPrefix+"/", 302)
		} else {
			loginPage(w, r, "Invalid login credentials")
		}
	} else {
		loginPage(w, r, "")
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		email := r.PostFormValue("email")
		u := &FullUser{Username: username, Email: email}
		u.SetPassword(password)
		if err := u.Add(); err != nil {
			_, column := dbu.Constrained(err)
			prompt := "Must be unique: " + column
			data := Profile{Prefix: pathPrefix, Prompt: prompt}
			data.Common.User = u.User() //currentUser(r)
			fmt.Println("PROMPT:", prompt)
			renderTemplate(w, r, "register", false, data)
			return
		}
		fmt.Println("USER", u)
		http.Redirect(w, r, pathPrefix+"/", 302)
	} else {
		data := Profile{Prefix: pathPrefix, Prompt: loginPrompt}
		data.Common.User = &User{}
		renderTemplate(w, r, "register", false, data)
	}
}

var webHandlers = []HFunc{
	{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/register", RegisterHandler},
	{"/login", LoginHandler},
	{"/logout", logoutPage},
	{"/user/list", usersListPage},
	{"/user/add", UserEdit},
	{"/user/edit/", UserEdit},
	{"/db/debug/", DebugPage},
	{"/ping", pingPage},
	{"/reload", reloadPage},
	//{"/password", PasswordReset},
	{"/", HomePage},
}
