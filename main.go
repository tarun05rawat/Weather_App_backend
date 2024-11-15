package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type WeatherData struct {
	City       string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Description string  `json:"description"`
	Clouds      int     `json:"clouds"`
	Humidity    int     `json:"humidity"`
	Pressure    int     `json:"pressure"`
}

type OpenWeatherMapResponse struct {
	Name   string `json:"name"`
	Main   struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
		Pressure int     `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Sys struct {
		Country string `json:"country"`
	} `json:"sys"`
}



func FetchWeather(city string) (WeatherData, error) {
	apiKey := os.Getenv("API_KEY")
	encodedCity := url.QueryEscape(city)
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", encodedCity, apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
			return WeatherData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return WeatherData{}, fmt.Errorf("OpenWeatherMap API error: %s", string(bodyBytes))
	}

	var apiResponse OpenWeatherMapResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
			return WeatherData{}, err
	}

	description := ""
	if len(apiResponse.Weather) > 0 {
			description = apiResponse.Weather[0].Description
	}

	weather := WeatherData{
			City:        apiResponse.Name,
			Temperature: apiResponse.Main.Temp,
			Description: description,
			Clouds:      apiResponse.Clouds.All,
			Humidity:    apiResponse.Main.Humidity,
			Pressure:    apiResponse.Main.Pressure,
	}
	return weather, nil
}


func WeatherHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	city := r.URL.Query().Get("city")
	if city == "" {
			http.Error(w, "City is required", http.StatusBadRequest)
			return
	}

	weatherData, err := FetchWeather(city)
	if err != nil {
			log.Printf("Error fetching weather data: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	json.NewEncoder(w).Encode(weatherData)
}


func main() {
	http.HandleFunc("/weather", WeatherHandler)
	

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
