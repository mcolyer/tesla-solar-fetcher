package main

import (
  "strconv"
  "encoding/json"
  "os"
  "bytes"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "time"
)
const OAUTH_CLIENT_ID  = "e4a9949fcfa04068f59abb5a658f2bac0a3428e4652315490b659d5ab3f35a9e"
const OAUTH_CLIENT_SECRET = "c75f14bbadc8bee3a7594412c31416f8300256d7668ea7e6e7f06727bfb9d220"
const USER_AGENT = "Mozilla/5.0 (Linux; Android 8.1.0; Pixel XL Build/OPM4.171019.021.D1; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/68.0.3440.91 Mobile Safari/537.36"
const APP_AGENT = "TeslaApp/3.4.4-350/fad4a582e/android/8.1.0"

type Authorization struct {
  AccessToken string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
  ExpiresIn int `json:"expires_in"`
  CreatedAt int `json:"created_at"`
}

func getToken(email string, password string) string {
  url := "https://owner-api.teslamotors.com/oauth/token"
  data := []byte(`{
    "grant_type": "password",
    "client_id": "`+OAUTH_CLIENT_ID+`",
    "client_secret": "`+OAUTH_CLIENT_SECRET+`",
    "email": "`+email+`",
    "password": "`+password+`"
  }`)

  req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
  if err != nil {
    log.Fatal("Error reading request. ", err)
  }

  // Set headers
  req.Header.Set("Content-Type", "application/json; charset=utf-8")
  req.Header.Set("User-Agent", USER_AGENT)
  req.Header.Set("X-Tesla-User-Agent", APP_AGENT)
  req.Header.Set("Host", "owner-api.teslamotors.com")

  // Set client timeout
  client := &http.Client{Timeout: time.Second * 10}

  // Send request
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal("Error reading response. ", err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal("Error reading body. ", err)
  }

  var authorization = new(Authorization)
  err = json.Unmarshal(body, &authorization)
  if(err != nil){
    fmt.Println("whoops:", err)
  }

  return authorization.AccessToken
}

type ProductsResponse struct {
  Count int `json:"count"`
  Products []Product `json:"response"`
}

type Product struct {
  Id string `json:"id"`
  EnergySiteId int `json:"energy_site_id"`
  ResourceType string `json:"resource_type"`
  SolarPower int `json:"expires_in"`
  SyncGridAlertEnabled bool `json:"sync_grid_alert_enabled"`
  BreakerAlertEnabled bool `json:"breaker_alert_enabled"`
}

func fetchEnergySite(token string) int {
  url := "https://owner-api.teslamotors.com/api/1/products"

  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    log.Fatal("Error reading request. ", err)
  }

  // Set headers
  req.Header.Set("Content-Type", "application/json; charset=utf-8")
  req.Header.Set("User-Agent", USER_AGENT)
  req.Header.Set("X-Tesla-User-Agent", APP_AGENT)
  req.Header.Set("Authorization", "Bearer "+token)
  req.Header.Set("Host", "owner-api.teslamotors.com")

  // Set client timeout
  client := &http.Client{Timeout: time.Second * 10}

  // Send request
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal("Error reading response. ", err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal("Error reading body. ", err)
  }

  parsed := new(ProductsResponse)
  err = json.Unmarshal(body, &parsed)
  if(err != nil){
    fmt.Println("whoops:", err)
  }

  return parsed.Products[0].EnergySiteId
}

type EnergyResponse struct {
  Unit UnitResponse `json:"response"`
}

type UnitResponse struct {
  SerialNumber string `json:"serial_number"`
  Samples []Sample `json:"time_series"`
}

type Sample struct {
  Timestamp string `json:"timestamp"`
  Solar int `json:"solar_power"`
  Battery int `json:"battery_power"`
  Grid int `json:"grid_power"`
}

func (this Sample) TimeAsNano() int64 {
  layout := "2006-01-02T15:04:00-07:00"
  t, err := time.Parse(layout, this.Timestamp)
  if err != nil {
          log.Fatal(err)
  }
  return t.UnixNano()
}

func fetchUsage(token string, site int, date string) []Sample {
  datetime := date+"T00:00:00.000Z"
  timezone := "America/Los_Angeles"
  url := "https://owner-api.teslamotors.com/api/1/energy_sites/"+strconv.Itoa(site)+"/history?kind=power&date="+datetime+"&period=_day&time_zone="+timezone

  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    log.Fatal("Error reading request. ", err)
  }

  // Set headers
  req.Header.Set("Content-Type", "application/json; charset=utf-8")
  req.Header.Set("User-Agent", USER_AGENT)
  req.Header.Set("X-Tesla-User-Agent", APP_AGENT)
  req.Header.Set("Authorization", "Bearer "+token)
  req.Header.Set("Host", "owner-api.teslamotors.com")

  // Set client timeout
  client := &http.Client{Timeout: time.Second * 10}

  // Send request
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal("Error reading response. ", err)
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal("Error reading body. ", err)
  }

  parsed := new(EnergyResponse)
  err = json.Unmarshal(body, &parsed)
  if(err != nil){
    fmt.Println("whoops:", err)
  }

  return parsed.Unit.Samples
}
func main() {
  email := os.Getenv("EMAIL")
  password := os.Getenv("PASSWORD")
  date := os.Getenv("DATE")
  if len(date) == 0 {
    date = time.Now().AddDate(0, 0, 0).Format("2006-01-02")
  }

  token := getToken(email, password)
  site := fetchEnergySite(token)
  samples := fetchUsage(token, site, date)

  for _, sample := range samples {
    fmt.Printf("solar_usage_wh value=%d %d\n", sample.Solar, sample.TimeAsNano())
  }
}
