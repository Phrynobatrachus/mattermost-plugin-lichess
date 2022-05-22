package main

type UserPrefs struct {
	Prefs    Prefs  `json:"prefs"`
	Language string `json:"language"`
}
