package vcsched

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/FogCreek/victor"
	"github.com/groob/radigast/plugins"
	"github.com/groob/vquery/axiom"
)

type Veracross struct {
	Query        string
	Username     string
	Password     string
	Client       string
	AllowedUsers []string
}

type schedule struct {
	First        string `json:"first_name"`
	Last         string `json:"last_name"`
	ScheduleView string `json:"schedule_view"`
}

func (a Veracross) Register() []victor.HandlerDocPair {
	// Allow everyone or just a specific group of users?
	var handler victor.HandlerFunc
	if len(a.AllowedUsers) == 0 {
		handler = a.scheduleFunc
	} else {
		handler = victor.OnlyAllow(a.AllowedUsers, a.scheduleFunc)
	}

	return []victor.HandlerDocPair{
		&victor.HandlerDoc{
			CmdHandler:     handler,
			CmdName:        "schedule",
			CmdDescription: "Student and Faculty Schedules",
			CmdUsage:       []string{"First Last"},
		},
	}
}

func (a Veracross) getSchedule() (*[]schedule, error) {
	var sched []schedule
	client, err := axiom.NewClient(a.Username, a.Password, a.Client)
	if err != nil {
		return &sched, err
	}
	req, err := http.NewRequest("POST", "https://axiom.veracross.com/"+a.Client+"/query/"+a.Query+"/result_data.json", nil)
	req.Header.Set("x-csrf-token", client.Token)
	resp, err := client.Do(req)
	if err != nil {
		return &sched, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&sched)
	if err != nil {
		return &sched, err
	}
	return &sched, err
}

func (a Veracross) scheduleFunc(s victor.State) {
	name := s.Fields()
	if len(name) < 2 {
		msg := "You must enter a first and last name as arguments"
		s.Chat().Send(s.Message().Channel().ID(), msg)
		return
	}

	schedules, err := a.getSchedule()
	if err != nil {
		log.Println(err)
		return
	}
	for _, schedule := range *schedules {
		if schedule.First != name[0] {
			continue
		}
		if schedule.Last != name[1] {
			continue
		}
		s.Chat().Send(s.Message().Channel().ID(), schedule.First+" "+schedule.Last+":\n"+schedule.ScheduleView)
		return
	}
}

func init() {
	plugins.Add("vcsched", func() plugins.Registrator {
		return &Veracross{}
	})
}
