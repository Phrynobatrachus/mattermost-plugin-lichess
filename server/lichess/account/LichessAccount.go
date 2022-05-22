package main

type LichessAccount struct {
	Id             string   `json:"id"`
	Username       string   `json:"username"`
	Online         bool     `json:"online"`
	Perfs          Perfs    `json:"perfs"`
	CreatedAt      int      `json:"createdAt"`
	Disabled       bool     `json:"disabled"`
	TosViolation   bool     `json:"tosViolation"`
	Profile        Profile  `json:"profile"`
	SeenAt         int      `json:"seenAt"`
	Patron         bool     `json:"patron"`
	Verified       bool     `json:"verified"`
	PlayTime       PlayTime `json:"playTime"`
	Title          string   `json:"title"`
	Url            string   `json:"url"`
	Playing        string   `json:"playing"`
	CompletionRate int      `json:"completionRate"`
	Count          Count    `json:"count"`
	Streaming      bool     `json:"streaming"`
	Followable     bool     `json:"followable"`
	Following      bool     `json:"following"`
	Blocking       bool     `json:"blocking"`
	FollowsYou     bool     `json:"followsYou"`
}
