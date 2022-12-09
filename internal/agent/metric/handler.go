package metric

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, bool) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, false
	}
	fv := v.Convert(floatType)
	return fv.Float(), true
}

func sendRequest(url, postData, method string) ([]byte, error) {
	var err error
	var answer []byte

	var res *http.Response
	defer res.Body.Close()

	switch strings.ToLower(method) {
	case "get":
		res, err = http.Get(url)
	case "post":
		res, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(postData)))

	default:
		log.Fatalf("wrong method: %s", method)
	}

	if err != nil {
		return answer, err
	}
	body, err := io.ReadAll(res.Body)

	answer = body

	return answer, err
}

func init() {
	rand.Seed(time.Now().Unix())
}
