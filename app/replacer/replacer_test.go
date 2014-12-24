package replacer

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"testing"
)

func TestSimpleReplace(t *testing.T) {
	inBuf := bytes.NewBuffer([]byte("http://common.melange:7776/js/query/1.1"))
	r := CreateReplacer(
		inBuf,
		`http://([a-z.]*)\.melange`,
		`http://$1.melange.xip.io`,
		`[^a-z\.]`,
	)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}

	expected := "http://common.melange.xip.io:7776/js/query/1.1"
	if string(data) != expected {
		t.Log("Data doesn't match.")
		t.Log("Expected", expected)
		t.Log("Received", string(data))
		t.Fail()
	}
}

func TestLongReplace(t *testing.T) {
	inBuf := bytes.NewBuffer([]byte("qtphasd;fnasdkjgnhttp://common.melange:7776/js/query/1.1"))
	r := CreateReplacer(
		inBuf,
		`http://([a-z\.]*)\.melange`,
		`http://$1.melange.xip.io`,
		`[^a-z\.]`,
	)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}

	expected := "qtphasd;fnasdkjgnhttp://common.melange.xip.io:7776/js/query/1.1"
	if string(data) != expected {
		t.Log("Data doesn't match.")
		t.Log("Expected", expected)
		t.Log("Received", string(data))
		t.Fail()
	}
}

func TestBorderReplace(t *testing.T) {
	inBuf := bytes.NewBuffer([]byte("qtphasd;fnashttp://common.melange:7776/js/query/1.1"))
	r := CreateReplacer(
		inBuf,
		`http://([a-z\.]*)\.melange`,
		`http://$1.melange.xip.io`,
		`[^a-z\.]`,
	)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}

	expected := "qtphasd;fnashttp://common.melange.xip.io:7776/js/query/1.1"
	if string(data) != expected {
		t.Log("Data doesn't match.")
		t.Log("Expected", expected)
		t.Log("Received", string(data))
		t.Fail()
	}
}

func BenchmarkReplace(t *testing.B) {
	for i := 0; i < t.N; i++ {
		randReader := io.LimitReader(rand.Reader, 1000*1000*5)
		r := CreateReplacer(
			randReader,
			`http://([a-z\.]*)\.melange`,
			`http://$1.melange.xip.io`,
			`[^a-z\.]`,
		)

		_, err := ioutil.ReadAll(r)
		if err != nil {
			t.Error(err)
		}
	}
}
