package main

import (
	"encoding/json"
	"spool-mock/config"
	"net/http"
)

// Return received msgids for refeeds
func refeed(w http.ResponseWriter, r *http.Request) {
	b, e := json.Marshal(config.RequeMsgids)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": true}`))
		return
	}

	w.Write(b)
}

func Http() {
	http.HandleFunc("/refeed", refeed)
    http.ListenAndServe("localhost:8432", nil)
}