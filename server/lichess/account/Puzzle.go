package lichess

type Puzzle struct {
	Games  int  `json:"games"`
	Rating int  `json:"rating"`
	Rd     int  `json:"rd"`
	Prog   int  `json:"prog"`
	Prov   bool `json:"prov"`
}
