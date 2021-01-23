/*
Copyright © 2020 Focus Centric inc. <dominicstpierre@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"staticbackendhq/cli/realtime"
	"staticbackendhq/cli/ws"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

type ctxvalue int

const (
	ctxStatus ctxvalue = iota
	ctxStart
	ctxPath
)

var (
	verbose bool
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a development server.",
	Long: fmt.Sprintf(`
%s

You may develop your application locally using the development server.

It has a direct mapping with StaticBackend API. You'll need no code changes 
when going from local to production.

There are some limitations that you can learn more about here.

%s
	`,
		clbold(clsecondary("StatickBackend development server")),
		clnote("https://staticbackend.com/cli"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("no-color").Value.String() == "true" {
			color.Disable()
		}

		verbose = cmd.Flag("no-log").Value.String() == "false"
		f := cmd.Flag("port")
		startServer(f.Value.String())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serverCmd.Flags().Int32P("port", "p", 8099, "dev server port")
	serverCmd.Flags().Bool("no-log", false, "prevents printing requests/responses info")
}

type devserver struct {
	db     map[string][]map[string]interface{}
	broker *realtime.Broker
}

type chainer func(h http.Handler) http.Handler

func chain(h http.Handler, middlewares ...chainer) http.Handler {
	next := h
	for _, m := range middlewares {
		next = m(next)
	}
	return next
}

func startServer(port string) {
	svr := &devserver{
		db: make(map[string][]map[string]interface{}),
	}

	// Start the Server Sent Event broker
	b := realtime.NewBroker(svr.findUser)

	svr.broker = b

	http.Handle("/register", chain(http.HandlerFunc(svr.register), svr.sb, svr.logger, svr.cors))
	http.Handle("/login", chain(http.HandlerFunc(svr.login), svr.sb, svr.logger, svr.cors))

	http.Handle("/db/", chain(http.HandlerFunc(svr.database), svr.sb, svr.logger, svr.cors))
	http.Handle("/query/", chain(http.HandlerFunc(svr.query), svr.sb, svr.logger, svr.cors))
	http.Handle("/sudo/", chain(http.HandlerFunc(svr.database), svr.sb, svr.reqRoot, svr.logger, svr.cors))
	http.Handle("/sudoquery/", chain(http.HandlerFunc(svr.query), svr.sb, svr.reqRoot, svr.logger, svr.cors))

	http.Handle("/postform/", chain(http.HandlerFunc(svr.postForm), svr.sb, svr.logger, svr.cors))

	http.Handle("/storage/upload", chain(http.HandlerFunc(svr.upload), svr.sb, svr.logger, svr.cors))
	http.Handle("/_servefile_/", chain(http.HandlerFunc(svr.serveFile), svr.logger, svr.cors))

	// WebSocket
	hub := ws.NewHub(svr.findUser)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	http.Handle("/sse/connect", chain(http.HandlerFunc(b.Accept), svr.logger, svr.cors))
	http.Handle("/sse/msg", chain(http.HandlerFunc(svr.message), svr.sb, svr.logger, svr.cors))

	// we create an admin
	users := make([]map[string]interface{}, 0)
	newUser := make(map[string]interface{})
	newUser["accountId"] = 1
	newUser["userId"] = 1
	newUser["email"] = "a@b.com"
	newUser["password"] = "test123"
	newUser["role"] = 100

	users = append(users, newUser)
	svr.db["sb_users"] = users

	fmt.Printf("%s http://localhost:%v\n", clsecondary("Server started at:"), port)
	fmt.Printf("%s: a@b.com with password test123\n", clsecondary("Your admin user is"))
	fmt.Println(http.ListenAndServe(":"+port, nil))
}

func (svr *devserver) register(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := svr.parse(r.Body, &data); err != nil {
		svr.respond(w, r, http.StatusBadRequest, err)
		return
	} else if data["email"] == nil {
		svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing email field"))
		return
	} else if data["password"] == nil {
		svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing password field"))
		return
	}

	users, ok := svr.db["sb_users"]
	if !ok {
		users = make([]map[string]interface{}, 0)

		data["accountId"] = 1
		data["userId"] = 1
	} else {
		newID := len(users) + 1

		data["accountId"] = newID
		data["userId"] = newID
	}

	data["role"] = 0

	users = append(users, data)

	svr.db["sb_users"] = users
	svr.respond(w, r, http.StatusOK, data["email"])
}

func (svr *devserver) login(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := svr.parse(r.Body, &data); err != nil {
		svr.respond(w, r, http.StatusBadRequest, err)
		return
	} else if data["email"] == nil {
		svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing email field"))
		return
	} else if data["password"] == nil {
		svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing password field"))
		return
	}

	user, err := svr.findUser(data["email"])
	if err != nil {
		svr.respond(w, r, http.StatusNotFound, err)
		return
	}

	if user["password"] != data["password"] {
		svr.respond(w, r, http.StatusNotFound, fmt.Errorf("invalid credential"))
		return
	}

	svr.respond(w, r, http.StatusOK, user["email"])
}

func (srv *devserver) respond(w http.ResponseWriter, r *http.Request, code int, v interface{}) {
	if err, ok := v.(error); ok {
		var tmp = new(struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		})
		tmp.Status = "error"
		tmp.Error = err.Error()
		v = tmp

		fmt.Printf("%s %s\n", clsecondary("error:"), cldanger(err.Error()))
	}

	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if verbose {
		ctx := r.Context()
		start, ok := ctx.Value(ctxStart).(time.Time)

		dur := time.Since(start)
		if ok {
			path, ok := ctx.Value(ctxPath).(string)
			if ok {
				fmt.Printf("%s\t%v\t%v\t%s\n",
					clsecondary(r.Method),
					clbold(code),
					clsecondary(dur),
					path,
				)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
}

func (svr *devserver) parse(r io.ReadCloser, v interface{}) error {
	defer r.Close()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return err
	}
	return nil
}

func (svr *devserver) findUser(email interface{}) (map[string]interface{}, error) {
	users, ok := svr.db["sb_users"]
	if !ok {
		return nil, fmt.Errorf("user %s not found", email)
	}

	for _, u := range users {
		if u["email"] == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", email)
}

func (svr *devserver) auth(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	key := r.Header.Get("Authorization")
	if len(key) == 0 {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/db/pub_") {
			return make(map[string]interface{})
		}
		http.Error(w, "need the Authorization HTTP header", http.StatusUnauthorized)
		return nil
	} else if strings.HasPrefix(key, "Bearer ") == false {
		http.Error(w, "need the Authorization HTTP header to be 'Bearer [user-token]'", http.StatusUnauthorized)
		return nil
	}

	email := strings.Replace(key, "Bearer ", "", -1)
	user, err := svr.findUser(email)
	if err != nil {
		http.Error(w, "unable to find this token, make sure the user has register first.", http.StatusUnauthorized)
		return nil
	}

	return user
}

type PagedResult struct {
	Page    int64                    `json:"page"`
	Size    int64                    `json:"size"`
	Total   int64                    `json:"total"`
	Results []map[string]interface{} `json:"results"`
}

func (svr *devserver) database(w http.ResponseWriter, r *http.Request) {
	user := svr.auth(w, r)
	if user == nil {
		return
	}

	col := ""
	_, r.URL.Path = svr.shiftPath(r.URL.Path)
	col, r.URL.Path = svr.shiftPath(r.URL.Path)

	if r.Method == http.MethodPost {
		var data map[string]interface{}
		if err := svr.parse(r.Body, &data); err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}
		data["id"] = svr.nextID(col)
		data["accountId"] = fmt.Sprintf("%v", user["accountId"])
		data["ownerId"] = fmt.Sprintf("%v", user["userId"])

		if err := svr.add(col, data); err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go svr.dbNotify(col, realtime.MsgTypeDBCreated, data)
		svr.respond(w, r, http.StatusCreated, data)
	} else if r.Method == http.MethodPut {
		id, _ := svr.shiftPath(r.URL.Path)
		if len(id) == 0 {
			svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing an id to the path: /db/[table]/[id]"))
			return
		}

		var data map[string]interface{}
		if err := svr.parse(r.Body, &data); err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}
		data["id"] = id
		data["accountId"] = fmt.Sprintf("%v", user["accountId"])

		if svr.ensureAccess(col, id, user, writePermission) == false {
			http.Error(w, "missing permission", http.StatusUnauthorized)
			return
		}

		orig, err := svr.fetch(col, id, func(v map[string]interface{}) bool {
			return true
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data["ownerId"] = orig["ownerId"]

		if err := svr.update(col, data); err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go svr.dbNotify(col, realtime.MsgTypeDBUpdated, data)
		svr.respond(w, r, http.StatusOK, true)
	} else if r.Method == http.MethodDelete {
		id, _ := svr.shiftPath(r.URL.Path)
		if len(id) == 0 {
			svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("missing an id to the path: /db/[table]/[id]"))
			return
		}

		fmt.Println(col, id, user)
		if svr.ensureAccess(col, id, user, writePermission) == false {
			http.Error(w, "missing permission", http.StatusUnauthorized)
			return
		}

		count, err := svr.del(col, id, fmt.Sprintf("%v", user["accountId"]))
		if err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		go svr.dbNotify(col, realtime.MsgTypeDBDeleted, id)
		svr.respond(w, r, http.StatusOK, count)
	} else if r.Method == http.MethodGet {
		id, _ := svr.shiftPath(r.URL.Path)
		if len(id) > 0 {
			if svr.ensureAccess(col, id, user, readPermission) == false {
				http.Error(w, "missing permission", http.StatusUnauthorized)
				return
			}

			rec, err := svr.fetch(col, id, func(v map[string]interface{}) bool { return true })
			if err != nil {
				svr.respond(w, r, http.StatusInternalServerError, err)
				return
			}
			svr.respond(w, r, http.StatusOK, rec)
			return
		}

		page, size := svr.getPagination(r.URL)
		list, total, err := svr.list(col, user, page, size)
		if err != nil {
			svr.respond(w, r, http.StatusInternalServerError, err)
			return
		}

		result := PagedResult{
			Page:    page,
			Size:    size,
			Total:   total,
			Results: list,
		}
		svr.respond(w, r, http.StatusOK, result)
	}
}

func (svr *devserver) dbNotify(col, typ string, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println("database notify: error converting to JSON", err)
		return
	}

	msg := realtime.Command{
		SID:     realtime.SystemID,
		Type:    typ,
		Data:    string(b),
		Channel: fmt.Sprintf("db-%s", col),
	}

	svr.broker.Publish(msg, msg.Channel)
}

func (svr *devserver) message(w http.ResponseWriter, r *http.Request) {
	var msg realtime.Command
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svr.broker.Broadcast <- msg

	svr.respond(w, r, http.StatusOK, true)
}

func (svr *devserver) shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

func (svr *devserver) add(col string, data map[string]interface{}) error {
	list, ok := svr.db[col]
	if !ok {
		list = make([]map[string]interface{}, 0)
	}

	list = append(list, data)

	svr.db[col] = list
	return nil
}

func (svr *devserver) update(col string, data map[string]interface{}) error {
	list, ok := svr.db[col]
	if !ok {
		return fmt.Errorf("this table %s is empty, cannot update your record.", col)
	}

	for idx, v := range list {
		if v["id"] == data["id"] {
			list[idx] = data
			svr.db[col] = list
			return nil
		}
	}
	return fmt.Errorf(
		"could not found the record with id %v and accountId %v in this table %s",
		data["id"],
		data["accountId"],
		col,
	)
}

func (svr *devserver) del(col string, id string, accountId interface{}) (int, error) {
	list, ok := svr.db[col]
	if !ok {
		return 0, nil
	}

	newList := make([]map[string]interface{}, 0)
	for _, v := range list {
		if v["id"] == id && v["accountId"] == accountId {
			continue
		}

		newList = append(newList, v)
	}

	svr.db[col] = newList
	return len(list) - len(newList), nil
}

func (svr *devserver) fetch(col, id string, cond func(doc map[string]interface{}) bool) (map[string]interface{}, error) {
	list, ok := svr.db[col]
	if !ok {
		fmt.Println("no repo")
		return nil, fmt.Errorf("your table %s is empty", col)
	}

	for _, v := range list {
		if v["id"] == id {
			if strings.HasPrefix(col, "pub_") {
				return v, nil
			} else if col == "sb_files" {
				// file are public by default
				return v, nil
			} else if found := cond(v); found {
				return v, nil
			} else {
				fmt.Println("not found")
				fmt.Println(v)
			}
		}
	}

	return nil, fmt.Errorf("could not find this id %s in the table %s", id, col)

}

func (svr *devserver) nextID(col string) string {
	list, ok := svr.db[col]
	if !ok {
		return "1"
	}
	return fmt.Sprintf("%d", len(list)+1)
}

func (svr *devserver) postForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := make(map[string]interface{})

	form := ""

	_, r.URL.Path = svr.shiftPath(r.URL.Path)
	form, r.URL.Path = svr.shiftPath(r.URL.Path)

	for key, val := range r.Form {
		data[key] = strings.Join(val, " ; ")
	}

	data["id"] = svr.nextID("sb_forms")
	data["form"] = form
	data["sb_posted"] = time.Now()

	if err := svr.add("sb_forms", data); err != nil {
		svr.respond(w, r, http.StatusInternalServerError, err)
		return
	}

	svr.respond(w, r, http.StatusOK, true)
}

func (svr *devserver) upload(w http.ResponseWriter, r *http.Request) {
	user := svr.auth(w, r)
	if user == nil {
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, h, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		svr.respond(w, r, http.StatusInternalServerError, err)
		return
	}

	id := svr.nextID("sb_files")

	rec := make(map[string]interface{})
	rec["id"] = id
	rec["accountId"] = user["accountId"]
	rec["file"] = b
	rec["ct"] = h.Header.Get("Content-Type")

	if err := svr.add("sb_files", rec); err != nil {
		svr.respond(w, r, http.StatusInternalServerError, err)
		return
	}

	svr.respond(w, r, http.StatusOK, fmt.Sprintf("/_servefile_/%v", rec["id"]))
}

func (svr *devserver) serveFile(w http.ResponseWriter, r *http.Request) {
	id := ""
	_, r.URL.Path = svr.shiftPath(r.URL.Path)
	id, r.URL.Path = svr.shiftPath(r.URL.Path)

	rec, err := svr.fetch("sb_files", id, nil)
	if err != nil {
		svr.respond(w, r, http.StatusInternalServerError, err)
		return
	}

	b, ok := rec["file"].([]byte)
	if !ok {
		svr.respond(w, r, http.StatusInternalServerError, fmt.Errorf("something is wrong with this file %s", id))
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("%v", rec["ct"]))
	w.Write(b)
}

func (svr *devserver) logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		ctx := context.WithValue(r.Context(), ctxStart, start)
		ctx = context.WithValue(ctx, ctxPath, path)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (svr *devserver) getPagination(u *url.URL) (page int64, size int64) {
	var err error

	page, err = strconv.ParseInt(u.Query().Get("page"), 10, 64)
	if err != nil {
		page = 1
	}

	size, err = strconv.ParseInt(u.Query().Get("size"), 10, 64)
	if err != nil {
		size = 25
	}

	return
}

func (svr *devserver) list(col string, user map[string]interface{}, page, size int64) ([]map[string]interface{}, int64, error) {
	list, ok := svr.db[col]
	if !ok {
		return nil, 0, fmt.Errorf("table %s does not exists", col)
	}

	filtered := make([]map[string]interface{}, 0)

	// if it's a public repo
	if strings.HasPrefix(col, "pub_") || user["role"] == 100 {
		filtered = list
	} else {
		rpl := readPermission(col)
		for _, rec := range list {
			switch rpl {
			case permGroup:
				if rec["accountId"] == fmt.Sprintf("%v", user["accountId"]) {
					filtered = append(filtered, rec)
				}
			case permOwner:
				if rec["accountId"] == fmt.Sprintf("%v", user["accountId"]) &&
					rec["ownerId"] == fmt.Sprintf("%v", user["userId"]) {
					filtered = append(filtered, rec)
				}
			default:
				// everyone can read based on the permission level
				filtered = append(filtered, rec)
			}
		}
	}

	sort.Sort(byID(filtered))

	skips := size * (page - 1)

	paged := make([]map[string]interface{}, 0)
	for idx, rec := range filtered {
		if int64(idx) < skips {
			continue
		} else if int64(idx)-skips > size {
			break
		}

		paged = append(paged, rec)
	}

	return paged, int64(len(filtered)), nil
}

type querycompare struct {
	op  string
	val interface{}
}

func (svr *devserver) query(w http.ResponseWriter, r *http.Request) {
	user := svr.auth(w, r)
	if user == nil {
		return
	}

	var col string
	_, r.URL.Path = svr.shiftPath(r.URL.Path)
	col, r.URL.Path = svr.shiftPath(r.URL.Path)

	list, ok := svr.db[col]
	if !ok {
		svr.respond(w, r, http.StatusNotFound, fmt.Errorf("this table %s does not exists", col))
		return
	}

	var clauses [][]interface{}
	if err := svr.parse(r.Body, &clauses); err != nil {
		svr.respond(w, r, http.StatusBadRequest, err)
		return
	}

	filter := make(map[string]querycompare)
	for i, clause := range clauses {
		if len(clause) != 3 {
			svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("The %d query clause did not contains the required 3 parameters (field, operator, value)", i+1))
			return
		}

		field, ok := clause[0].(string)
		if !ok {
			svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("The %d query clause's field parameter must be a string: %v", i+1, clause[0]))
			return
		}

		op, ok := clause[1].(string)
		if !ok {
			svr.respond(w, r, http.StatusBadRequest, fmt.Errorf("The %d query clause's operator must be a string: %v", i+1, clause[1]))
			return
		}

		switch op {
		case "=", "==":
			filter[field] = querycompare{op: "=", val: clause[2]}
		case "!=", "<>":
			filter[field] = querycompare{op: "!", val: clause[2]}
		default:
			fmt.Printf("%s %v %s %v %s\n", cldanger("The query's"), i, cldanger("operator"), op, cldanger("is not currently supported in the dev server."))
		}
	}

	if strings.HasPrefix(col, "pub_") == false {
		filtered := make([]map[string]interface{}, 0)

		rpl := readPermission(col)
		for _, rec := range list {
			switch rpl {
			case permGroup:
				if rec["accountId"] == fmt.Sprintf("%v", user["accountId"]) {
					filtered = append(filtered, rec)
				}
			case permOwner:
				if rec["accountId"] == fmt.Sprintf("%v", user["accountId"]) &&
					rec["ownerId"] == fmt.Sprintf("%v", user["userId"]) {
					filtered = append(filtered, rec)
				}
			default:
				// everyone can read based on the permission level
				filtered = append(filtered, rec)
			}
		}

		list = filtered
	}

	page, size := svr.getPagination(r.URL)

	skips := size * (page - 1)

	result := PagedResult{
		Page: page,
		Size: size,
	}

	filtered := make([]map[string]interface{}, 0)
	valid := true
	for _, rec := range list {
		valid = true
		for k, v := range filter {
			if v.op == "=" {
				if fmt.Sprintf("%v", rec[k]) != fmt.Sprintf("%v", v.val) {
					valid = false
					break
				}
			} else if v.op == "!" {
				if fmt.Sprintf("%v", rec[k]) == fmt.Sprintf("%v", v.val) {
					valid = false
					break
				}
			}
		}

		if valid {
			filtered = append(filtered, rec)
		}
	}

	result.Total = int64(len(filtered))

	sort.Sort(byIDDesc(filtered))

	paged := make([]map[string]interface{}, 0)
	for idx, rec := range filtered {
		if int64(idx) < skips {
			continue
		} else if int64(idx)-skips > size {
			break
		}

		paged = append(paged, rec)
	}

	result.Results = paged

	svr.respond(w, r, http.StatusOK, result)
}

func (svr *devserver) sb(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header.Get("SB-PUBLIC-KEY")) == 0 {
			fmt.Println("A HTTP header named SB-PUBLIC-KEY is required.")
			fmt.Println("In development mode you may pass any value for instance: SB-PUBLIC-KEY: my-key.")
			fmt.Println("You'll receive this key when you create your account.")
			svr.respond(w, r, http.StatusUnauthorized, fmt.Errorf("you need to supply a SB-PUBLIC-KEY header."))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (svr *devserver) reqRoot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Authorization")
		if len(key) == 0 {
			http.Error(w, "need the Authorization HTTP header", http.StatusUnauthorized)
			return
		} else if strings.HasPrefix(key, "Bearer ") == false {
			http.Error(w, "need the Authorization HTTP header to be 'Bearer [root-token]'", http.StatusUnauthorized)
			return
		}

		email := strings.Replace(key, "Bearer ", "", -1)
		user, err := svr.findUser(email)
		if err != nil {
			http.Error(w, "unable to find this token, make sure the user has register first.", http.StatusUnauthorized)
			return
		} else if user["role"] != 100 {
			http.Error(w, "user missing permission", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (svr *devserver) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		origin := r.Header.Get("Origin")

		// Always set Vary headers
		// see https://github.com/rs/cors/issues/10,
		//     https://github.com/rs/cors/commit/dbdca4d95feaa7511a46e6f1efb3b3aa505bc43f#commitcomment-12352001
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")

		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		headers.Set("Access-Control-Allow-Origin", origin)
		// Spec says: Since the list of methods can be unbounded, simply returning the method indicated
		// by Access-Control-Request-Method (if supported) can be enough
		headers.Set("Access-Control-Allow-Methods", strings.ToUpper(r.Header.Get("Access-Control-Request-Method")))

		headers.Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type permissionLevel int

const (
	permOwner permissionLevel = iota
	permGroup
	permEveryone
)

func getPermission(repo string) (owner string, group string, everyone string) {
	// default permission
	owner, group, everyone = "7", "4", "0"

	re := regexp.MustCompile(`_\d\d\d_$`)
	if re.MatchString(repo) == false {
		return
	}

	results := re.FindAllString(repo, -1)
	if len(results) != 1 {
		return
	}

	perm := strings.Replace(results[0], "_", "", -1)

	if len(perm) != 3 {
		return
	}

	owner = string(perm[0])
	group = string(perm[1])
	everyone = string(perm[2])
	return
}

func writePermission(repo string) permissionLevel {
	_, g, e := getPermission(repo)

	if canWrite(e) {
		return permEveryone
	}
	if canWrite(g) {
		return permGroup
	}
	return permOwner
}

func readPermission(repo string) permissionLevel {
	if strings.HasPrefix(repo, "pub_") {
		return permEveryone
	}

	_, g, e := getPermission(repo)

	if canRead(e) {
		return permEveryone
	}
	if canRead(g) {
		return permGroup
	}
	return permOwner
}

func canWrite(s string) bool {
	i, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return uint8(i)&uint8(2) != 0
}

func canRead(s string) bool {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println(err)
	}
	return uint8(i)&uint8(4) != 0
}

func (svr *devserver) ensureAccess(repo, id string, user map[string]interface{}, p func(repo string) permissionLevel) bool {
	if user["role"] == 100 {
		return true
	}

	cond := func(v map[string]interface{}) bool {
		return true
	}

	switch p(repo) {
	case permGroup:
		cond = func(v map[string]interface{}) bool {
			a1 := fmt.Sprintf("%v", v["accountId"])
			a2 := fmt.Sprintf("%v", user["accountId"])
			return a1 == a2
		}
	case permOwner:
		cond = func(v map[string]interface{}) bool {
			a1 := fmt.Sprintf("%v", v["accountId"])
			a2 := fmt.Sprintf("%v", user["accountId"])
			o1 := fmt.Sprintf("%v", v["ownerId"])
			o2 := fmt.Sprintf("%v", user["userId"])
			return a1 == a2 && o1 == o2
		}
	}

	if _, err := svr.fetch(repo, id, cond); err != nil {
		fmt.Println("fetch err: ", err)
		return false
	}

	return true
}
