package common

import (
	"math/rand"
	"time"
)

var blessings = []string{
	"祝老板天天开心，笑口常开",
	"老板好运加持，喜上眉梢",
	"老板开心指数爆表，心情超好",
	"老板福运连连，事事顺心",
	"祝老板笑容常在，阳光满满",
	"老板星光加冕，魅力值拉满",
	"祝一切顺利，心想事成",
	"赞赞赞，掌声送给老板",
	"老板今天一定大吉大利",
	"老板快乐环绕，活力满满",
	"老板喜气腾腾，步步高升",
	"老板元气满满，笑容满满",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomBlessing() string {
	return blessings[rand.Intn(len(blessings))]
}
