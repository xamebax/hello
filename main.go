package main

import (
  "encoding/json"
  "net/http"
  "strings"
  "log"
  "time"
  "fmt"
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

type weatherUnderground struct {
  apiKey string
}

type openWeatherMap struct {}

type weatherProvider interface {
  temperature(city string) (float64, error)
}

type multiWeatherProvider []weatherProvider

// How are errors handled? => do they have to be function arguments?
// "we return a non-nil error to the client, who's expected to deal it in a way that makes sense in the calling context"

func (w openWeatherMap) temperature(city string) (float64, error) {
  resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=5aaaf27195261fc52ff9ba54b2e6d557&q=" + city)
  if err != nil {
    return 0, err
  }

  // Go's defer statement schedules a function call (the deferred function) to be run immediately before the function executing the defer returns.
  defer resp.Body.Close() // this will close the body once the function runs

  var d struct {
    Main struct {
      Kelvin float64 `json:"temp"`
    } `json:"main"`
  }

  // `err` can be defined a line above and this if block will work, but
  // the err variable won't be limited to the scope of the `if` block
  if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
    return 0, err
  }

  log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)

  return d.Main.Kelvin, nil
}

func (w weatherUnderground) temperature(city string) (float64, error) {
  resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/" + city + ".json")
  if err !=nil {
    return 0, err
  }

  defer resp.Body.Close()

  var d struct {
    Observation struct {
      Celsius float64 `json:"temp_c"`
    } `json:"current_observation"`
  }

  if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
    return 0, err
  }

  kelvin := d.Observation.Celsius + 273.15
  log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
  return kelvin, nil
}

// This can be passed anywhere that accepts a weatherProvider:
func (w multiWeatherProvider) temperature(city string) (float64, error) {
  sum := 0.0

  for _, provider := range w {
    k, err := provider.temperature(city)
    if err != nil {
      return 0, err
    }

    sum += k
  }

  return sum / float64(len(w)), nil
}


func temperature(city string, providers ...weatherProvider) (float64, error) {
  sum := 0.0

  for _, provider := range providers {
    k, err := provider.temperature(city)
    if err != nil {
      return 0, err
    }

    sum += k
  }

  return sum / float64(len(providers)), nil
}



func hello(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("hello!")) //converts byte to string
}


func main() {

  mw := multiWeatherProvider{
    openWeatherMap{},
    weatherUnderground{apiKey: "14d238fb55d3edd9"},
  }

  http.HandleFunc("/hello", hello)

  // defining the handler inline; everything after "/weather/" => city
  http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
    begin := time.Now()
    city := strings.SplitN(r.URL.Path, "/", 3)[2]

    temp, err := mw.temperature(city)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(map[string]interface{}{
      "city": city,
      "temp": temp,
      "took": time.Since(begin).String(),
    })
  })

  http.ListenAndServe(":8080", nil)
}
