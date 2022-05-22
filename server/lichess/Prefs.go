package lichess

type Prefs struct {
	Dark          bool   `json:"dark"`
	Transp        bool   `json:"transp"`
	BgImg         string `json:"bgImg"`
	Is3d          bool   `json:"is3d"`
	Theme         string `json:"theme"`
	PieceSet      string `json:"pieceSet"`
	Theme3d       string `json:"theme3d"`
	PieceSet3d    string `json:"pieceSet3d"`
	SoundSet      string `json:"soundSet"`
	Blindfold     int    `json:"blindfold"`
	AutoQueen     int    `json:"autoQueen"`
	AutoThreefold int    `json:"autoThreefold"`
	Takeback      int    `json:"takeback"`
	Moretime      int    `json:"moretime"`
	ClockTenths   int    `json:"clockTenths"`
	ClockBar      bool   `json:"clockBar"`
	ClockSound    bool   `json:"clockSound"`
	Premove       bool   `json:"premove"`
	Animation     int    `json:"animation"`
	Captured      bool   `json:"captured"`
	Follow        bool   `json:"follow"`
	Highlight     bool   `json:"highlight"`
	Destination   bool   `json:"destination"`
	Coords        int    `json:"coords"`
	Replay        int    `json:"replay"`
	Challenge     int    `json:"challenge"`
	Message       int    `json:"message"`
	CoordColor    int    `json:"coordColor"`
	SubmitMove    int    `json:"submitMove"`
	ConfirmResign int    `json:"confirmResign"`
	InsightShare  int    `json:"insightShare"`
	KeyboardMove  int    `json:"keyboardMove"`
	Zen           int    `json:"zen"`
	MoveEvent     int    `json:"moveEvent"`
	RookCastle    int    `json:"rookCastle"`
}
