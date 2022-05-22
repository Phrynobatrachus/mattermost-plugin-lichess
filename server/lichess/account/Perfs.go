package lichess

type Perfs struct {
	Chess960       Chess960       `json:"chess960"`
	Atomic         Atomic         `json:"atomic"`
	RacingKings    RacingKings    `json:"racingKings"`
	UltraBullet    UltraBullet    `json:"ultraBullet"`
	Blitz          Blitz          `json:"blitz"`
	KingOfTheHill  KingOfTheHill  `json:"kingOfTheHill"`
	Bullet         Bullet         `json:"bullet"`
	Correspondence Correspondence `json:"correspondence"`
	Horde          Horde          `json:"horde"`
	Puzzle         Puzzle         `json:"puzzle"`
	Classical      Classical      `json:"classical"`
	Rapid          Rapid          `json:"rapid"`
	Storm          Storm          `json:"storm"`
}
