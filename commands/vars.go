package commands

import (
	"qaz/logger"
	"qaz/repo"
	stks "qaz/stacks"
	"sync"
)

var (
	config  Config
	stacks  map[string]*stks.Stack
	region  string
	project string
	wg      sync.WaitGroup
	gitrepo repo.Repo
	log     = logger.Logger{
		DebugMode: &run.debug,
	}
)

// config environment variable
const (
	configENV     = "QAZ_CONFIG"
	defaultconfig = "config.yml"
)

// run.var used as a central point for command data
var run = struct {
	cfgSource  string
	tplSource  string
	profile    string
	tplSources []string
	stacks     map[string]string
	all        bool
	version    bool
	request    string
	debug      bool
	funcEvent  string
	changeName string
	stackName  string
	rollback   bool
	colors     bool
	cfgRaw     string
	gituser    string
	gitpass    string
	gitrsa     string
}{}
