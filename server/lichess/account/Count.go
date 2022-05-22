package main

type Count struct {
	All      int `json:"all"`
	Rated    int `json:"rated"`
	Ai       int `json:"ai"`
	Draw     int `json:"draw"`
	DrawH    int `json:"drawH"`
	Loss     int `json:"loss"`
	LossH    int `json:"lossH"`
	Win      int `json:"win"`
	WinH     int `json:"winH"`
	Bookmark int `json:"bookmark"`
	Playing  int `json:"playing"`
	Import   int `json:"import"`
	Me       int `json:"me"`
}
