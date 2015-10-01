package vcrotation

import (
	"log"
	"strings"
	"time"

	"github.com/FogCreek/victor"
	"github.com/groob/radigast/plugins"
	"github.com/whitby/vcapi"
)

type VeracrossAPI struct {
	Username     string
	Password     string
	Client       string
	AllowedUsers []string
}

func (a VeracrossAPI) Register() []victor.HandlerDocPair {
	var handler victor.HandlerFunc
	if len(a.AllowedUsers) == 0 {
		handler = a.rotationFunc
	} else {
		handler = victor.OnlyAllow(a.AllowedUsers, a.rotationFunc)
	}
	return []victor.HandlerDocPair{
		&victor.HandlerDoc{
			CmdHandler:     handler,
			CmdName:        "rotation",
			CmdDescription: "Upper School Rotation",
			CmdUsage:       []string{"2015-09-30"},
		},
	}
}

func (a VeracrossAPI) getRotation(date *time.Time) (*vcapi.RotationDays, error) {
	config := &vcapi.Config{
		Username:   a.Username, // API Username
		Password:   a.Password, // API Password
		SchoolID:   a.Client,   // Client, school name
		APIVersion: "v2",
	}
	from := date.Format(vcapi.VCTimeFormat)
	to := date.Add(24 * time.Hour).Format(vcapi.VCTimeFormat)
	client := vcapi.NewClient(config)
	opt := &vcapi.ListOptions{
		Params: vcapi.Params{
			"date_from": from,
			"date_to":   to,
		}}
	day, err := client.RotationDays.List(opt)
	if err != nil {
		return nil, err
	}
	return &day[0], nil
}
func (a VeracrossAPI) rotationFunc(s victor.State) {
	var date time.Time
	var err error
	input := strings.Join(s.Fields(), " ")
	switch input {
	case "today":
		date = time.Now()
	case "tomorrow":
		date = time.Now().Add(24 * time.Hour)
	default:
		date, err = time.Parse(vcapi.VCTimeFormat, input)
		if err != nil {
			log.Println(err)
		}
	}
	rotation, err := a.getRotation(&date)
	if err != nil {
		log.Println(err)
	}
	s.Chat().Send(s.Message().Channel().ID(), rotation.Description)
}

func init() {
	plugins.Add("vcrotation", func() plugins.Registrator {
		return &VeracrossAPI{}
	})
}
