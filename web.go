package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	gorilla "github.com/gorilla/handlers"
)

const (
	logDir    = "logs"
	accessLog = "access.log"
	errorLog  = "error.log"
	uCookie   = "userinfo"

	//disable caching for testing
	//maxAge	  = 259200
	maxAge = 0
)

var (
	errorFile    *os.File
	tmpl         map[string]*template.Template
	TDir         = "assets/templates"
	BaseFile     = "base.html"
	fm           = template.FuncMap{"isTrue": isTrue}
	Verbose      = false
	cacheControl = fmt.Sprintf("public, max-age=%d", maxAge)
)

type HFunc struct {
	Path string
	Func http.HandlerFunc
}

func auditLog(uid int64, ip, action, msg string) {
	fmt.Println(uid, ip, strings.ToLower(action), msg)
}

func RemoteHost(r *http.Request) string {
	remote_addr := r.Header.Get("X-Forwarded-For")
	if len(remote_addr) == 0 {
		var err error
		remote_addr, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "REMOTE ADDR ERR:", err)
		}
	}
	if len(remote_addr) > 0 && remote_addr[0] == ':' {
		remote_addr = MyIp()
	}
	return remote_addr
}

func MyIp() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !strings.HasPrefix(ipnet.String(), "127.") && strings.Index(ipnet.String(), ":") == -1 {
			return strings.Split(ipnet.String(), "/")[0]
		}
	}
	return ""
}

func LoadTemplates(dir, base string, funcMap template.FuncMap) {
	tmpl = make(map[string]*template.Template)
	files, err := filepath.Glob(path.Join(dir, "*.html"))
	if err != nil {
		panic(err)
	}
	full := path.Join(dir, base)
	for _, file := range files {
		name := filepath.Base(file)
		if name == base {
			continue
		}
		if Verbose {
			fmt.Println("COMPILE: ", name)
		}
		t := template.New(name).Funcs(funcMap)
		tmpl[name] = template.Must(t.ParseFiles(file, full))
	}
}

// helper for layout, example for other helpers
func isTrue(in interface{}) string {
	yes, err := strconv.ParseBool(in.(string))
	if err != nil {
		return err.Error()
	}
	if yes {
		return "true"
	}
	return "false"
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tname string, base bool, data interface{}) {
	name := string(tname + ".html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var err error
	if base {
		err = tmpl[name].ExecuteTemplate(w, "base", data)
	} else {
		err = tmpl[name].Execute(w, data)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// for loading an object from an http post
func objFromForm(obj interface{}, values map[string][]string) {
	val := reflect.ValueOf(obj)
	base := reflect.Indirect(val)
	t := reflect.TypeOf(base.Interface())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		b := base.Field(i)
		if val, ok := values[f.Name]; ok {
			switch b.Interface().(type) {
			case string:
				b.SetString(val[0])
			case int:
				i, _ := strconv.Atoi(val[0])
				b.SetInt(int64(i))
			case int64:
				i, _ := strconv.ParseInt(val[0], 0, 64)
				b.SetInt(i)
			case uint:
				i, _ := strconv.ParseUint(val[0], 0, 64)
				b.SetUint(i)
			case uint32:
				i, _ := strconv.ParseUint(val[0], 0, 32)
				b.SetUint(i)
			default:
				fmt.Println("unhandled field type for:", f.Name, "type:", b.Type())
			}
		}
	}
}

func currentUser(r *http.Request) *User {
	cookie, err := r.Cookie(uCookie)
	if err != nil {
		return nil
	}
	return UserFromCookie(cookie.Value)
}

func reloadPage(w http.ResponseWriter, r *http.Request) {
	LoadTemplates(TDir, BaseFile, fm)
	fmt.Fprintln(w, "reloaded")
}

func FaviconPage(w http.ResponseWriter, r *http.Request) {
	fav := filepath.Join(assets_dir, "static/images/favicon.ico")
	http.ServeFile(w, r, fav)
}

func init() {
	assets_dir, _ = filepath.Abs(assets_dir) // CWD may change at runtime
	TDir, _ = filepath.Abs(TDir)             // CWD may change at runtime
	LoadKey(secretKey)
}

func StaticPage(w http.ResponseWriter, r *http.Request) {
	const skip = len(pathPrefix)
	name := filepath.Join(assets_dir, r.URL.Path[skip:])
	file, err := os.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	w.Header().Set("Cache-control", cacheControl)
	http.ServeContent(w, r, name, fi.ModTime(), file)
}

func ErrorLog(r *http.Request, msg string, args ...interface{}) {
	user := currentUser(r)
	remote_addr := RemoteHost(r)
	fmt.Fprintln(errorFile, time.Now().Format(log_layout), remote_addr, user.ID, msg, args)
}

func ProtectedPage(r *http.Request) bool {
	pages := []string{"/register", "/login", "/static"}
	const pref = len(pathPrefix)
	if len(r.URL.Path) >= pref {
		path := r.URL.Path[pref:]
		plen := len(path)
		for _, p := range pages {
			l := len(p)
			if plen >= l && path[:l] == p {
				return false
			}
		}
	}
	return true
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if ProtectedPage(r) {
				const u = pathPrefix + "/login"
				user := currentUser(r)
				if user == nil {
					http.Redirect(w, r, u, 302)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// to have a common root to run against a a proxy
// automatically redirects non-proxy link
func usePrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		http.Redirect(w, r, pathPrefix+r.URL.Path, 307)
	} else {
		http.Redirect(w, r, pathPrefix+r.URL.Path, 302)
	}
}

func webServer(handlers []HFunc) {
	LoadTemplates(TDir, BaseFile, fm)
	http.HandleFunc("/", usePrefix)
	for _, h := range handlers {
		http.Handle(pathPrefix+h.Path, authMiddleware(h.Func))
	}

	http_server := fmt.Sprintf("%s:%d", MyIp(), http_port)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Panic(err)
	}
	logFile := filepath.Join(logDir, accessLog)
	accessLog, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Panic("Error opening access log:", err)
	}
	errorPath := filepath.Join(logDir, "error.log")
	errorFile, err = os.OpenFile(errorPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Panic("Error opening error log:", err)
	}

	fmt.Println("serve up web:", http_server)
	err = http.ListenAndServe(http_server, gorilla.CompressHandler(gorilla.LoggingHandler(accessLog, http.DefaultServeMux)))
	if err != nil {
		panic(err)
	}
}
