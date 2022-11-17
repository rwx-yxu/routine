package routine

import (
	"encoding/json"
	"fmt"
	"log"

	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/conf"
	"github.com/rwxrob/help"
	"github.com/rwxrob/vars"
	"gopkg.in/yaml.v3"
)

func init() {
	Z.Conf.SoftInit()
	Z.Vars.SoftInit()
}

var Cmd = &Z.Cmd{

	Name:      `routine`,
	Summary:   `a command that prints out current day routine schedule`,
	Version:   `v0.4.1`,
	Copyright: `Copyright 2022 Yongle Xu`,
	License:   `Apache-2.0`,
	Site:      `yonglexu.dev`,
	Source:    `git@github.com:rwx-yxu/routine.git`,
	Issues:    `github.com/rwx-yxu/routine/issues`,

	Commands: []*Z.Cmd{
		printCmd,
		// standard external branch imports (see rwxrob/{help,conf,vars})
		help.Cmd, conf.Cmd, vars.Cmd,
	},

	// Add custom BonzaiMark template extensions (or overwrite existing ones).

	Description: `
		{{cmd .Name}} is a tool that prints out google calendar events for the day along with any reminders. The output will be to a text file that is fed into a thermal printer. Google API key will be required to be set using routine conf edit.
			`,
}

var printCmd = &Z.Cmd{
	Name:     `daily`,
	Summary:  `print current to standard output (default)`,
	Commands: []*Z.Cmd{help.Cmd},
	Call: func(x *Z.Cmd, _ ...string) error {
		val, _ := Z.Conf.Data()
		if val == "" {
			log.Printf("API key has not been initialized. Please set the API key by using routine conf edit")
			return nil
		}

		jsn, _ := conf2Json([]byte(val))

		fmt.Println(string(jsn))

		return nil
	},
}

func conf2Json(body []byte) ([]byte, error) {
	m := make(map[string]any)
	err := yaml.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	js, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return []byte(js), err
}
