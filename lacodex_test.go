package lacodex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/konkers/lacodex/model"

	"github.com/stretchr/testify/assert"

	"github.com/phayes/freeport"
)

type testLC struct {
	l          *LaCodex
	exitC      chan struct{}
	imageField string
}

func newTestLC(t *testing.T) *testLC {
	dbFile, err := ioutil.TempFile("", "*.db")
	assert.NoError(t, err, "Can't get tempFile")
	defer os.Remove(dbFile.Name())

	port, err := freeport.GetFreePort()
	assert.NoError(t, err, "Can't get free port")
	host := "localhost:" + strconv.Itoa(port)

	l, err := NewLaCodex(&Config{
		DbPath:     dbFile.Name(),
		ListenAddr: host,
	})
	assert.NoError(t, err, "Can't create new LaCodex")

	exitC := make(chan struct{})
	go func() {
		l.Run()
		close(exitC)
	}()
	return &testLC{
		l:          l,
		exitC:      exitC,
		imageField: "image",
	}
}

func (tlc *testLC) Shutdown() {
	tlc.l.Shutdown()
	<-tlc.exitC
}

func (tlc *testLC) PutImage(t *testing.T, filename string) int {
	file, err := os.Open(filename)
	assert.NoError(t, err, "Can't open %s", filename)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(tlc.imageField, filepath.Base(filename))
	assert.NoError(t, err, "Can't create part")

	_, err = io.Copy(part, file)
	assert.NoError(t, err, "Can't copy file part")

	err = writer.Close()
	assert.NoError(t, err, "Can't close writer")

	url := fmt.Sprintf("http://%s/image/upload", tlc.l.config.ListenAddr)
	req, err := http.NewRequest("PUT", url, body)
	assert.NoError(t, err, "Can't create new req")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err, "Can't process request")
	resp.Body.Close()
	return resp.StatusCode
}

func (tlc *testLC) GetImages(t *testing.T) []*model.ImageMetadata {
	url := fmt.Sprintf("http://%s/image/list", tlc.l.config.ListenAddr)
	r := testGet(t, url)
	var imgs []*model.ImageMetadata
	err := json.Unmarshal([]byte(r), &imgs)
	assert.NoError(t, err, "Can't decode json: %s", r)
	return imgs
}

func (tlc *testLC) GetRecords(t *testing.T) []*model.Record {
	url := fmt.Sprintf("http://%s/record/list", tlc.l.config.ListenAddr)
	r := testGet(t, url)
	var imgs []*model.Record
	err := json.Unmarshal([]byte(r), &imgs)
	assert.NoError(t, err, "Can't decode json: %s", r)
	return imgs
}

func testTimeParse(t *testing.T, value string) time.Time {
	v, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", value)
	assert.NoError(t, err, "Can't parse time")
	return v
}

func TestLaCodexShutdown(t *testing.T) {
	tlc := newTestLC(t)
	tlc.Shutdown()
}

func TestBadAddr(t *testing.T) {
	l := newTestLC(t)
	l.l.config.ListenAddr = "-1"
	err := l.l.Run()
	assert.NoError(t, err, "Expected error from ListendAddr -1")
}

func TestLaCodexBadPerms(t *testing.T) {
	dbFile, err := ioutil.TempFile("", "*.db")
	assert.NoError(t, err, "Can't get tempFile")
	defer os.Remove(dbFile.Name())

	os.Chmod(dbFile.Name(), 0)

	port, err := freeport.GetFreePort()
	assert.NoError(t, err, "Can't get free port")
	host := "localhost:" + strconv.Itoa(port)

	_, err = NewLaCodex(&Config{
		DbPath:     dbFile.Name(),
		ListenAddr: host,
	})

	assert.Error(t, err, "NewLaCodex didn't fail on bad database file")
}

func TestSameContentFileUpload(t *testing.T) {
	tlc := newTestLC(t)
	tlc.PutImage(t, "testdata/screenshots/230700_20190519134140_1.png")
	tlc.PutImage(t, "testdata/screenshots/230700_20190519134145_1.png")

	imgs := tlc.GetImages(t)
	assert.Equal(t, []*model.ImageMetadata{
		&model.ImageMetadata{
			Id:         1,
			Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
			CapturedAt: testTimeParse(t, "2019-05-19 13:41:40 -0700 PDT"),
			FileName:   "230700_20190519134140_1.png",
			Record:     1,
		},
		&model.ImageMetadata{
			Id:         2,
			Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
			CapturedAt: testTimeParse(t, "2019-05-19 13:41:45 -0700 PDT"),
			FileName:   "230700_20190519134145_1.png",
			Record:     2,
		},
	}, imgs)

	tlc.Shutdown()
}

func TestSameFileUpload(t *testing.T) {
	tlc := newTestLC(t)
	tlc.PutImage(t, "testdata/screenshots/230700_20190519134140_1.png")
	tlc.PutImage(t, "testdata/screenshots/230700_20190519134140_1.png")

	imgs := tlc.GetImages(t)
	assert.Equal(t, []*model.ImageMetadata{
		&model.ImageMetadata{
			Id:         1,
			Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
			CapturedAt: testTimeParse(t, "2019-05-19 13:41:40 -0700 PDT"),
			FileName:   "230700_20190519134140_1.png",
			Record:     1,
		},
	}, imgs)

	tlc.Shutdown()
}

func TestWrongFieldnameFileUpload(t *testing.T) {
	tlc := newTestLC(t)
	tlc.imageField = "nope"
	status := tlc.PutImage(t, "testdata/screenshots/230700_20190519134140_1.png")
	assert.NotEqual(t, http.StatusOK, status)
	tlc.Shutdown()
}

func TestBadFileUpload(t *testing.T) {
	tlc := newTestLC(t)

	status := tlc.PutImage(t, "testdata/bad_images/undecodable.png")
	assert.NotEqual(t, http.StatusOK, status)

	tlc.Shutdown()
}

func TestMapImageUpload(t *testing.T) {
	tlc := newTestLC(t)
	tlc.PutImage(t, "testdata/screenshots/230700_20190517185334_1.png")
	tlc.PutImage(t, "testdata/screenshots/230700_20190519134140_1.png")

	imgs := tlc.GetImages(t)
	assert.Equal(t, []*model.ImageMetadata{
		&model.ImageMetadata{
			Id:         1,
			Hash:       "sha256-c1d85db281056ffb2f43214219dafba0aba67032d7b05fae393d4c1d0f22fe59",
			CapturedAt: testTimeParse(t, "2019-05-17 18:53:34 -0700 PDT"),
			FileName:   "230700_20190517185334_1.png",
			Record:     0,
		},
		&model.ImageMetadata{
			Id:         2,
			Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
			CapturedAt: testTimeParse(t, "2019-05-19 13:41:40 -0700 PDT"),
			FileName:   "230700_20190519134140_1.png",
			Record:     1,
		},
	}, imgs)

	recs := tlc.GetRecords(t)
	assert.Equal(t,
		[]*model.Record{
			&model.Record{
				Id:         1,
				Type:       2,
				Text:       "Offer 3 lights to the heavens.\nOK\ni",
				Subject:    "",
				Index:      nil,
				Keyphrases: map[model.KeyphraseType][]string{},
			},
		}, recs)

	tlc.Shutdown()
}
