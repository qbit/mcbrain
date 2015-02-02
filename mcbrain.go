package main

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"html/template"
	"log"
	"net/http"
)

const mcTemplate = `
<html>
<head>
<title>MCBrain</title>
<script src="//code.jquery.com/jquery-2.1.3.min.js"></script>
<script src="//cdn.datatables.net/1.10.4/js/jquery.dataTables.min.js"></script>
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap-theme.min.css">
<link rel="stylesheet" href="//cdn.datatables.net/1.10.4/css/jquery.dataTables.min.css">
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/js/bootstrap.min.js"></script>
<style>
</style>
</head>
<body>
<table>
<thead>
<td>Word</td>
<td>Classification</td>
<td>Value</td>
</thead>
{{range $key, $value := .}}
<tr>
<td class="word">{{$key}}</td>
<td class="wordc"></td>
<td>{{$value}}</td>
</tr>
{{end}}
</table>
<script>
var words = $('.word'), i, l, cats = $('.wordc');

for (i = 0, l = words.length; i < l; i++) {
  var s = $(words[i]).text().split("____")
  var w = s[0];
  var c = s[1];
  $(words[i]).text(w);
  $(cats[i]).text(c);
}
$('table').DataTable();
</script>
</body>
</html>`

var store = "classifier_bayes_words_twss"
var templ = template.Must(template.New("mcbrain").Parse(mcTemplate))

type twssMap map[string]int

func getData() (twssMap, error) {
	data := twssMap{}
	client, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return nil, err
	}

	defer client.Close()

	keys := client.Cmd("HKEYS", store)
	keyStr, err := keys.List()
	if err != nil {
		return nil, err
	}

	for key := range keyStr {
		v := client.Cmd("HGET", store, keyStr[key])
		data[keyStr[key]], _ = v.Int()
	}

	return data, nil
}

func brainDisplay(w http.ResponseWriter, req *http.Request) {
	data, err := getData()
	if err != nil {
		log.Printf("Can't get data: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	err = templ.Execute(w, data)
	if err != nil {
		log.Printf("template execution failed! %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func brainJSON(w http.ResponseWriter, req *http.Request) {
	data, err := getData()
	if err != nil {
		log.Printf("can't get data: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ohgod!, %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func main() {
	http.HandleFunc("/", brainDisplay)
	http.HandleFunc("/json", brainJSON)
	http.ListenAndServe(":3011", nil)
}
