package main

import (
	"encoding/json"
	"expvar"
	"net/http"
)

func (app *application) statsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte("{"))
	appendComma := false

	expvar.Do(func(kv expvar.KeyValue) {
		if kv.Key == "memstats" || kv.Key == "cmdline" {
			return // Skip memstats and cmdline
		}
		if appendComma {
			w.Write([]byte(","))
		} else {
			appendComma = true
		}

		b, _ := json.Marshal(kv.Key)
		w.Write(b)
		w.Write([]byte(":"))
		w.Write([]byte(kv.Value.String())) // here .String() returns json compatible string so, no marshal

	})
	w.Write([]byte("}"))
}
