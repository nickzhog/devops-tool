package metric

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
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
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	request, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewBuffer([]byte(postData)))
	if err != nil {
		return []byte(``), err
	}

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return []byte(``), err
	}
	defer res.Body.Close()

	answer, err := io.ReadAll(res.Body)

	return answer, err
}
