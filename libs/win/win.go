package win

import (
	"github.com/ezydark/force/libs/win/admin"
	"github.com/ezydark/force/libs/win/fs"
	"github.com/shirou/gopsutil/process"
)

var Admin *admin.Admin
var Fs *fs.Fs
var Process *process.Process
