package metric

import (
	"io"
	"math/rand"
	"net/http"
	"reflect"
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

func sendRequest(url, postData string) ([]byte, error) {
	var err error
	var answer []byte
	res, err := http.Post(url, "text/plain", nil)

	if err != nil {
		return answer, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	answer = body

	return answer, err
}

func init() {
	rand.Seed(time.Now().Unix())
}
