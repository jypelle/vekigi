package device

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jypelle/vekigi/apimodel"
	"github.com/jypelle/vekigi/internal/srv/config"
	"github.com/jypelle/vekigi/internal/srv/event"
	"github.com/jypelle/vekigi/internal/tool"
	"github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type Api struct {
	lock         sync.RWMutex
	eventChannel chan event.ApiEvent

	router    *mux.Router
	apiRouter *mux.Router
	server    *http.Server

	config  *config.ServerConfig
	askDone chan bool
	done    chan bool
}

func NewApi(config *config.ServerConfig) *Api {
	api := Api{
		config:       config,
		eventChannel: make(chan event.ApiEvent),
		askDone:      make(chan bool),
		done:         make(chan bool),
	}

	api.router = mux.NewRouter().Schemes("https").Subrouter()
	api.router = mux.NewRouter().StrictSlash(false)

	// API Routes
	api.apiRouter = api.router.PathPrefix("/api").Subrouter()
	api.apiRouter.NotFoundHandler = http.HandlerFunc(ErrorNotFoundAction)
	api.apiRouter.MethodNotAllowedHandler = http.HandlerFunc(ErrorMethodNotAllowedAction)

	// Auth middleware
	api.apiRouter.Use(
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if rec := recover(); rec != nil {
						logrus.Warningf("recovered from panic : [%v] - stack trace : \n [%s]", rec, debug.Stack())
						strMessage := fmt.Sprintf("%v", rec)
						GlobalErrorAction(w, strMessage, http.StatusInternalServerError)
					}
				}()

				// Check API Key
				apiKey := r.Header.Get("x-api-key")
				if apiKey != config.ServerParam.ApiParam.ApiKey {
					ErrorStatusAction(w, r, http.StatusForbidden)
					return
				}

				logrus.Debugf("PATH: %s %s", r.Host, r.URL.Path)

				handler.ServeHTTP(w, r)
			})
		})

	// API Routes

	// Create server check endpoint
	api.apiRouter.HandleFunc("/is_alive",
		func(w http.ResponseWriter, r *http.Request) {
			ErrorStatusAction(w, r, http.StatusOK)
		}).Methods("GET")
	api.apiRouter.HandleFunc("/webradio/play/{group_id}/{index_id}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			groupIdstr, ok := vars["group_id"]
			if !ok {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			groupId, err := strconv.ParseInt(groupIdstr, 10, 0)
			if err != nil {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}

			indexIdstr, ok := vars["index_id"]
			if !ok {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			indexId, err := strconv.ParseInt(indexIdstr, 10, 0)
			if err != nil {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}

			result := make(chan error)
			api.eventChannel <- event.ApiEvent{Result: result, Data: event.ApiEventWebradioPlayData{WebradioId: apimodel.WebradioId{GroupId: groupId, IndexId: indexId}}}
			err = <-result
			if err == nil {
				ErrorStatusAction(w, r, http.StatusOK)
			} else {
				GlobalErrorAction(w, err.Error(), http.StatusForbidden)
			}
		}).Methods("POST")
	api.apiRouter.HandleFunc("/playlist/play/{playlist_id}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			playlistIdstr, ok := vars["playlist_id"]
			if !ok {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			playlistId, err := strconv.ParseInt(playlistIdstr, 10, 0)
			if err != nil {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			result := make(chan error)
			api.eventChannel <- event.ApiEvent{Result: result, Data: event.ApiEventPlaylistPlayData{PlaylistId: apimodel.PlaylistId(playlistId)}}
			err = <-result
			if err == nil {
				ErrorStatusAction(w, r, http.StatusOK)
			} else {
				GlobalErrorAction(w, err.Error(), http.StatusForbidden)
			}
		}).Methods("POST")
	api.apiRouter.HandleFunc("/audio/volume/{volume}",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			volumeStr, ok := vars["volume"]
			if !ok {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			volume, err := strconv.ParseInt(volumeStr, 10, 0)
			if err != nil {
				ErrorStatusAction(w, r, http.StatusBadRequest)
				return
			}
			result := make(chan error)
			api.eventChannel <- event.ApiEvent{Result: result, Data: event.ApiEventAudioVolumeData{Volume: volume}}
			err = <-result
			if err == nil {
				ErrorStatusAction(w, r, http.StatusOK)
			} else {
				GlobalErrorAction(w, err.Error(), http.StatusForbidden)
			}
		}).Methods("POST")

	// Tell the browser that it's OK for JS to communicate with the server
	headersOk := handlers.AllowedHeaders([]string{"Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	api.server = &http.Server{
		Addr:         ":" + strconv.FormatInt(config.ServerParam.ApiParam.SslPort, 10),
		Handler:      handlers.CompressHandler(handlers.CORS(originsOk, headersOk, methodsOk)(api.router)),
		ReadTimeout:  time.Second * 240,
		WriteTimeout: time.Second * 240,
		IdleTimeout:  time.Second * 240,
	}

	return &api
}

func (d *Api) Start() {
	logrus.Infof("Start api device")

	existServerCert, err := tool.IsFileExists(d.selfSignedCertFilename())
	if err != nil {
		logrus.Fatalf("Unable to access %s: %v\n", d.selfSignedCertFilename(), err)
	}

	existServerKey, err := tool.IsFileExists(d.selfSignedKeyFilename())
	if err != nil {
		logrus.Fatalf("Unable to access %s: %v\n", d.selfSignedKeyFilename(), err)
	}

	if !existServerCert || !existServerKey {
		logrus.Info("Missing cert and key files, trying to generate them...")
		err = tool.GenerateTlsCertificate(
			"jypelle",
			"Vekigi Server",
			d.selfSignedKeyFilename(),
			d.selfSignedCertFilename(),
			[]string{})
		if err != nil {
			logrus.Fatalf("Unable to generate cert and key files : %v\n", err)
		}
		logrus.Info("Self-signed cert and key files generated")
	}

	// Launch https server
	go func() {
		err := d.server.ListenAndServeTLS(d.selfSignedCertFilename(), d.selfSignedKeyFilename())
		if err != nil && err.Error() != "http: Server closed" {
			logrus.Error(err)
		}
	}()
}

func (d *Api) StopSendingEvent() {
	logrus.Infof("Stop api device")
	d.server.Shutdown(context.Background())
	//close(d.eventChannel)
}

func (d *Api) EventChannel() chan event.ApiEvent {
	return d.eventChannel
}

func (d *Api) selfSignedKeyFilename() string {
	return filepath.Join(d.config.ConfigDir, "key.pem")
}

func (d *Api) selfSignedCertFilename() string {
	return filepath.Join(d.config.ConfigDir, "cert.pem")
}

func ErrorNotFoundAction(w http.ResponseWriter, r *http.Request) {
	ErrorStatusAction(w, r, http.StatusNotFound)
}

func ErrorMethodNotAllowedAction(w http.ResponseWriter, r *http.Request) {
	ErrorStatusAction(w, r, http.StatusMethodNotAllowed)
}

func ErrorStatusAction(w http.ResponseWriter, r *http.Request, status int) {
	ErrorMessageAction(w, "", status)
}

func GlobalErrorAction(w http.ResponseWriter, message string, status int) {
	ErrorMessageAction(w, message, status)
}

func ErrorMessageAction(w http.ResponseWriter, title string, status int) {
	errorMessage := &apimodel.ErrorMessage{
		ErrStatusCode: status,
		ErrMessage:    title,
	}

	if title == "" {
		switch status {
		case http.StatusOK:
			errorMessage.ErrMessage = "Ok"
		case http.StatusNotFound:
			errorMessage.ErrMessage = "Page not found"
		case http.StatusMethodNotAllowed:
			errorMessage.ErrMessage = "Method not allowed"
		case http.StatusForbidden:
			errorMessage.ErrMessage = "Forbidden"
		case http.StatusServiceUnavailable:
			errorMessage.ErrMessage = "Service unavailable"
		case http.StatusBadRequest:
			errorMessage.ErrMessage = "Bad request"
		default:
			errorMessage.ErrMessage = "Internal error"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorMessage)
}
