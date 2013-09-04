package ex01

import (
	"appengine"
	"appengine/datastore"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func init() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/cmd/register", signUpHandler)
	http.HandleFunc("/cmd/jange", signInHandler)
	http.HandleFunc("/cmd/nakseo", nakseoHandler)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	indexPage := "/page/index.html"
	code := http.StatusFound
	http.Redirect(w, r, indexPage, code)
}

// register, POST
func signUpHandler(w http.ResponseWriter, r *http.Request) {
	if "POST" != r.Method {
		return
	}

	if err := r.ParseForm(); nil != err {
		log.Println(err)
		return
	}

	id := r.Form.Get("idStr")
	pw := r.Form.Get("password")
	c := appengine.NewContext(r)

	// 키 생성
	// NewKey
	// params: context, kind, stringId, intId, parent
	// return: key
	encKey := datastore.NewKey(c, "jange", id, 0, nil)

	jange := &Jange{
		IdStr:    id,
		Password: pw,
		EncKey:   encKey,
	}

	// Put
	// params: context, key, src (interface {})
	// return: key, err
	_, err := datastore.Put(c, encKey, jange)
	if nil != err {
		log.Println(err)
		return
	}

	res := make(map[string]string)
	res["idStr"] = id
	res["encKey"] = encKey.Encode()
	res["result"] = "success"

	data, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, string(data))
}

// jange, POST
func signInHandler(w http.ResponseWriter, r *http.Request) {
	if "POST" != r.Method {
		return
	}

	if err := r.ParseForm(); nil != err {
		log.Println(err)
		return
	}

	id := r.Form.Get("idStr")
	pw := r.Form.Get("password")
	c := appengine.NewContext(r)

	// Get
	// params: context, key, dst (interface {})
	// return: err
	encKey := datastore.NewKey(c, "jange", id, 0, nil)

	jange := &Jange{}

	if err := datastore.Get(c, encKey, jange); err != nil {
		log.Println(err)
		return
	}

	res := make(map[string]string)
	res["idStr"] = id
	res["encKey"] = encKey.Encode()

	if pw == jange.Password {
		res["result"] = "success"
	}

	data, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, string(data))
}

func nakseoHandler(w http.ResponseWriter, r *http.Request) {
	// nakseo
	if err := r.ParseForm(); nil != err {
		log.Println(err)
		return
	}

	switch r.Method {
	case "POST":
		addNakseo(w, r)
	case "GET":
		getNakseo(w, r)
	case "PUT":
		putNakseo(w, r)
	case "DELETE":
		delNakseo(w, r)
	}
}

func getHashOfContent(content string) string {
	hash := sha256.New()
	hash.Write([]byte(content))
	md := hash.Sum(nil)
	return base64.URLEncoding.EncodeToString(md)
}

func addNakseo(w http.ResponseWriter, r *http.Request) {
	encUserKey := r.Form.Get("encKey")
	content := r.Form.Get("content")

	if "" == content {
		return
	}

	c := appengine.NewContext(r)
	jange := &Jange{}

	// func DecodeKey(encoded string) (*Key, error)
	userKey, err := datastore.DecodeKey(encUserKey)

	if nil != err {
		log.Println(err)
		return
	}

	// 유저 아이디 얻기
	if err := datastore.Get(c, userKey, jange); err != nil {
		log.Println(err)
		return
	}

	regDate := time.Now().String()
	hash := getHashOfContent(regDate)

	// userKey: parent key
	nakseoKey := datastore.NewKey(c, "nakseo", hash, 0, userKey)

	nakseo := &Nakseo{
		EncKey:      nakseoKey,
		Content:     content,
		Owner:       jange.IdStr,
		EncOwnerKey: userKey,
		RegDate:     regDate,
	}

	if _, err := datastore.Put(c, nakseoKey, nakseo); err != nil {
		log.Println(err)
		return
	}

	res := make(map[string]string)
	res["result"] = "success"

	data, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, string(data))
}

func getNakseo(w http.ResponseWriter, r *http.Request) {
	encUserKey := r.Form.Get("encKey")
	nextToken := r.Form.Get("nextToken")
	fetch := r.Form.Get("fetch")
	next := 0

	userKey, err := datastore.DecodeKey(encUserKey)

	if nil != err {
		log.Println(err)
		return
	}

	// 쿼리 생성
	// func NewQuery(kind string) *Query
	q := datastore.NewQuery("nakseo").Order("-RegDate")

	if "all" != fetch {
		q = q.Ancestor(userKey)
	}

	q = q.Offset(next)

	max := 10

	c := appengine.NewContext(r)

	var cursor datastore.Cursor
	if "" != nextToken {
		cursor, _ = datastore.DecodeCursor(nextToken)
		q = q.Start(cursor)
	}

	iter := q.Run(c)
	nakseoList := make([]*Nakseo, max)

	count := 0
	for i := 0; i < max; i++ {
		nakseo := &Nakseo{}
		_, err := iter.Next(nakseo)

		if err == datastore.Done {
			break
		}

		if err != nil {
			return
		}

		cursor, _ = iter.Cursor()
		nakseoList[i] = nakseo
		count = i + 1
	}

	res := &NakseoResult{
		Result:    "success",
		Nakseo:    nakseoList[:count],
		NextToken: cursor.String(),
	}

	if 0 == count {
		res.Result = "no_more"
	}

	data, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, string(data))
}

func putNakseo(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
}

func delNakseo(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL)
}
