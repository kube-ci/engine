package workplan_viewer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/appscode/go/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kube-ci/engine/apis/engine/v1alpha1"
	"github.com/kube-ci/engine/pkg/logs"
	"k8s.io/client-go/rest"
)

const (
	statusPath = "/namespaces/%s/workplans/%s"
	logsPath   = "/namespaces/%s/workplans/%s/steps/%s"
)

var logController *logs.LogController

func wsWriter(ws *websocket.Conn, reader io.Reader) {
	defer func() {
		ws.Close()
	}()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
			return
		}
	}
}

func getWebsocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// if auth successful upgrade connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	log.Infoln("Websocket Connected")
	return ws, nil
}

// TODO: remove
func serveUpgrade(w http.ResponseWriter, r *http.Request) {
	log.Infoln("Serving websocket template...")
	wsTemplate := template.Must(template.New("ws-template").Parse(wsTemplate))
	wsTemplate.Execute(w, "ws://"+r.Host+r.URL.Path)
}

func serveLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := logs.Query{
		Namespace: vars["namespace"],
		Workplan:  vars["workplan"],
		Step:      vars["step"],
	}

	if !websocket.IsWebSocketUpgrade(r) { // TODO: remove
		serveUpgrade(w, r)
		return
	}

	ws, err := getWebsocket(w, r)
	if err != nil {
		log.Errorf("can't create websocket, reason: %s", err)
		return
	}
	defer func() {
		if err != nil {
			log.Errorln(err)
			ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			ws.Close()
		}
	}()

	reader, err := logController.LogReader(query)
	if err != nil {
		err = fmt.Errorf("can't get LogReader, reason: %s", err)
		return
	}

	go wsWriter(ws, reader)
}

func serveWorkplan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := logs.Query{
		Namespace: vars["namespace"],
		Workplan:  vars["workplan"],
	}

	var err error
	defer func() {
		if err != nil {
			log.Errorln(err)
			w.Write([]byte(err.Error()))
		}
	}()

	// list all steps in the given Workplan along with their status
	workplanStatus, err := logController.WorkplanStatus(query)
	if err != nil {
		err = fmt.Errorf("can't get WorkplanStatus, reason: %s", err)
		return
	}

	// add log link with status
	for i, steps := range workplanStatus.StepTree {
		for j, step := range steps {
			if step.Status == v1alpha1.ContainerRunning || step.Status == v1alpha1.ContainerTerminated {
				logPath := fmt.Sprintf(logsPath, query.Namespace, query.Workplan, step.Name)
				statusWithLogLink := fmt.Sprintf(`%s <a href=%s>[Logs]</a>`, step.Status, logPath)
				workplanStatus.StepTree[i][j].Status = v1alpha1.ContainerStatus(statusWithLogLink)
			}
		}
	}

	statusJson, err := json.Marshal(workplanStatus)
	if err != nil {
		err = fmt.Errorf("can't marshal WorkplanStatus to json, reason: %s", err)
		return
	}

	statusTemplate := template.Must(template.New("status-template").Parse(statusTemplate))
	statusTemplate.Execute(w, string(statusJson))
}

func Serve(clientConfig *rest.Config) (err error) {
	logController, err = logs.NewLogController(clientConfig)
	if err != nil {
		return fmt.Errorf("error initializing log-controller, reason: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf(statusPath, "{namespace}", "{workplan}"), serveWorkplan)
	r.HandleFunc(fmt.Sprintf(logsPath, "{namespace}", "{workplan}", "{step}"), serveLog)
	http.Handle("/", r)

	fmt.Println("Starting workplan-viewer...")
	return http.ListenAndServe(":9090", nil)
}

var wsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Workplan Logs</title>
    <style>
        body {
            background-color: #272727;
        }
        .log {
            margin: 20px;
            font-family: monospace;
            font-size: 16px;
            color: #22da26;;
        }
    </style>
</head>

<body>
<div class="log" id="log"/>
<script type="text/javascript">
    (function() {
        var log = document.getElementById("log");
        var conn = new WebSocket("{{.}}");
        conn.onmessage = function(evt) {
            log.innerHTML = log.innerHTML + '<br>' + evt.data;
        }
    })();
</script>
</body>
</html>`

var statusTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Workplan Status</title>
    <meta http-equiv="refresh" content="30">
</head>

<body>
<script>
    var obj = JSON.parse({{.}});
    var str = JSON.stringify(obj, undefined, 4);
    document.body.appendChild(document.createElement('pre')).innerHTML = str;
</script>
</body>

</html>
`
