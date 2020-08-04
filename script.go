package libIM

import (
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

var ScriptHeader = `#!/bin/bash
#$ -e $0.e
#$ -o $0.o
`

func CreateShell(fileName, script string, args ...string) {
	var file = osUtil.Create(fileName)
	defer simpleUtil.DeferClose(file)

	fmtUtil.Fprintf(file, "%ssh %s %s\n", ScriptHeader, script, strings.Join(args, " "))
}
