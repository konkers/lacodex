package lacodex

import (
	"context"
	"encoding/json"
	"image"
	"io"
	"net/http"

	"github.com/cskr/pubsub"

	"github.com/go-zoo/bone"

	"github.com/golang/glog"
	"github.com/konkers/lacodex/ingest"
	"github.com/konkers/lacodex/model"

	"github.com/asdine/storm"
	"github.com/konkers/lacodex/imagedb"
)

// Config contains the configuration for LaCodex.
type Config struct {
	DbPath     string `json:"db"`
	ListenAddr string `json:"listen"`
}

// LaCodex is an instance of LaCodex.
type LaCodex struct {
	config  *Config
	db      *storm.DB
	idb     *imagedb.ImageDB
	records storm.Node

	ps       *pubsub.PubSub
	shutdown chan struct{}
}

// NewLaCodex creates a new LaCodex instance.
func NewLaCodex(config *Config) (*LaCodex, error) {
	db, err := storm.Open(config.DbPath)
	if err != nil {
		return nil, err
	}

	idb := imagedb.NewImageDB(db.From("imagedb"))

	return &LaCodex{
		config:   config,
		db:       db,
		idb:      idb,
		records:  db.From("records"),
		ps:       pubsub.New(0),
		shutdown: make(chan struct{}),
	}, nil
}

func (l *LaCodex) addImage(img image.Image, fileName string) error {
	if meta, _ := l.idb.LookupFile(fileName); meta != nil {
		glog.V(2).Infof("already have %s", fileName)
		return nil
	}

	glog.Infof("adding %s", fileName)
	gameImg := ingest.CropGameImage(img)
	record, err := ingest.IngestImage(gameImg)
	glog.Infof("%#v %v", record, err)
	if err == nil {
		err = l.records.Save(record)
		if err != nil {
			return err
		}
	} else {
		record = &model.Record{Id: 0}
	}

	err = l.idb.ImportScreenshot(fileName, record.Id, gameImg)
	if err != nil {
		return err
	}
	l.ps.Pub(nil, "update")

	return nil
}

func (l *LaCodex) imageUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("image")
	if err != nil {
		httpError(w, http.StatusBadRequest, "No image file in request: %v", err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		httpError(w, http.StatusBadRequest, "Error decoding image: %v", err)
		return
	}

	err = l.addImage(img, handler.Filename)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Error adding image: %v", err)
		return
	}
}

func (l *LaCodex) listImages(w io.Writer) error {
	meta, err := l.idb.ListImages()
	if err != nil {
		return err
	}
	json.NewEncoder(w).Encode(meta)
	return nil
}

func (l *LaCodex) listRecords(w io.Writer) error {
	var records []*model.Record
	err := l.records.All(&records)
	if err != nil {
		return err
	}
	json.NewEncoder(w).Encode(records)
	return nil
}

func (l *LaCodex) Run() error {
	mux := bone.New()

	mux.Put("/image/upload", http.HandlerFunc(l.imageUploadHandler))
	mux.Get("/image/list", WsHandler(l.ps, l.listImages))
	mux.Get("/record/list", WsHandler(l.ps, l.listRecords))

	glog.Infof("Serving at http://%s/ ...", l.config.ListenAddr)

	var srv http.Server
	srv.Handler = mux
	srv.Addr = l.config.ListenAddr
	go func() {
		<-l.shutdown
		err := srv.Shutdown(context.Background())
		warnIfError(err, "HTTP server Shutdown")
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		glog.Warningf("HTTP server ListenAndServe: %v", err)
	}

	l.ps.Pub(nil, "exit")

	return nil
}

func (l *LaCodex) Shutdown() {
	close(l.shutdown)
}
