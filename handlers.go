package main

import (
	"fmt"
	"net/http"
	"time"

	"html/template"
)

var definedOps = []OP{
	OP{
		ID:          `info`,
		Description: "Update NetScaler Info",
		Path:        "update/info",
		LastUpdate:  "Never",
	},
	OP{
		ID:          `mappings`,
		Description: "Update NetScaler Mappings",
		Path:        "update/mappings",
		LastUpdate:  "Never",
	},
}

// OP holds the info needed to perform an operation against a defined handler.
type OP struct {
	ID          string
	Description string
	Path        string
	LastUpdate  string
}

func (a *API) opsHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New(`opsPage`)
	tmpl, err := t.Parse(opsPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","error":"` + err.Error() + `"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	ops := make([]OP, 0, len(definedOps))
	for _, o := range definedOps {
		O := OP{
			ID:          o.ID,
			Description: o.Description,
			Path:        o.Path,
			LastUpdate:  o.LastUpdate,
		}
		switch o.ID {
		case `info`:
			switch {
			case a.lastInfo != time.Time{}:
				O.LastUpdate = a.lastInfo.String()
			}
		case `mappings`:
			switch {
			case a.lastMapping != time.Time{}:
				O.LastUpdate = a.lastMapping.String()
			}
		}
		ops = append(ops, O)
	}
	tmpl.Execute(w, ops)
}

func (a *API) updateMappingsHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New(`statusPage`)
	tmpl, err := t.Parse(statusPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","error":"` + err.Error() + `"}`))
		return
	}
	tmp := struct {
		Status   string
		Reason   string
		Previous string
	}{
		Status:   `request sent`,
		Reason:   `N/A`,
		Previous: `N/A`,
	}
	switch {
	case a.mappingFB.good():
		a.lastMapping = time.Now()
		go pools.collectMappings(nil, true)
		go flipAfter(a.mappingFB, time.Minute*60)
	default:
		tmp.Status = `request not sent`
		tmp.Reason = `request too soon after previous request`
		tmp.Previous = a.lastMapping.String()
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, tmp)
}

func (a *API) collectInfoHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New(`statusPage`)
	tmpl, err := t.Parse(statusPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","error":"` + err.Error() + `"}`))
		return
	}
	tmp := struct {
		Status   string
		Reason   string
		Previous string
	}{
		Status:   `request sent`,
		Reason:   `N/A`,
		Previous: `N/A`,
	}
	switch {
	case a.infoFB.good():
		a.lastInfo = time.Now()
		go pools.collectNSInfo()
		go flipAfter(a.infoFB, time.Minute*60)
	default:
		tmp.Status = `request not sent`
		tmp.Reason = `request too soon after previous request`
		tmp.Previous = a.lastInfo.String()
	}
	w.WriteHeader(http.StatusOK)
	fmt.Println("ERR:", tmpl.Execute(w, tmp))
}

func flipAfter(FB *FlipBit, dur time.Duration) {
	timer := time.NewTimer(dur)
	<-timer.C
	FB.flip()
}

var opsPage = `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Netscaler Exporter Ops</title>
	<style>
		table {
		  font-family: arial, sans-serif;
		  border-collapse: collapse;
		  width: 100%;
		}
		td, th {
		  border: 1px solid #dddddd;
		  text-align: left;
		  padding: 8px;
		}
		tr:nth-child(even) {
		  background-color: #dddddd;
		}
	</style>
</head>
<body>
	<table>
		<tr>
			<th>Operation</th>
			<th>Request</th>
			<th>Last Request</th>
		</tr>
		{{ range . }}
			<tr>
				<td>{{ .Description }}</td>
				<td><a href="{{ .Path }}">Generate</a></td>
				<td>{{ .LastUpdate }}</td>
			</tr>
		{{ end }}
    </table>
</body>
</html>`

var statusPage = `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Netscaler Exporter Ops</title>
	<style>
		table {
		  font-family: arial, sans-serif;
		  border-collapse: collapse;
		  width: 100%;
		}
		td, th {
		  border: 1px solid #dddddd;
		  text-align: left;
		  padding: 8px;
		}
		tr:nth-child(even) {
		  background-color: #dddddd;
		}
	</style>
</head>
<body>
	<table>
		<tr>
			<th>Status</th>
			<th>Reason</th>
			<th>Previous Request</th>
		</tr>
			<tr>
				<td>{{ .Status }}</td>
				<td>{{ .Reason }}</td>
				<td>{{ .Previous }}</td>
			</tr>
    </table>
</body>
</html>`
