package main

import (
  "encoding/json"
  "net/http"
  "strings"
)

// func HandleFunc(pattern string, handler func(ResponseWriter, *Request))
// func ListenAndServe(addr string, handler Handler) error

// tag: kind of like metadata; allows to use the encoding/json package.
type weatherData struct {
  Name string `json:"name"`
  Main struct {
    Kelvin float64 `json:"temp"`
  } `json:"main"`
}

// How are errors handled? => do they have to be function arguments?
// "we return a non-nil error to the client, who's expected to deal it in a way that makes sense in the calling context"
//         input         output
func query(city string) (weatherData, error) {
  resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=5aaaf27195261fc52ff9ba54b2e6d557&q=" + city)
  if err != nil {
    return weatherData{}, err
  }

  // Go's defer statement schedules a function call (the deferred function) to be run immediately before the function executing the defer returns.
  defer resp.Body.Close() // this will close the body once the function runs

  var d weatherData

  // `err` can be defined a line above and this if block will work, but
  // the err variable won't be limited to the scope of the `if` block
  if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
    return weatherData{}, err
  }

  return d, nil
}


func hello(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("hello!")) //converts byte to string
}



func main() {
  http.HandleFunc("/hello", hello)

    // defining the handler inline; everything after "/weather/" => city
  http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
    city := strings.SplitN(r.URL.Path, "/", 3)[2]

    data, err := query(city)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(data)
  })

  http.ListenAndServe(":8080", nil)
}
