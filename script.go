package libIM

import (
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

var ScriptHeader = `#!/bin/bash
set -e
#$ -cwd
`

var ScriptFooter = `if [ "$?" != "0" ]; then
exit 100
fi
`

func CreateShell(fileName, script string, args ...string) {
	var file = osUtil.Create(fileName)
	defer simpleUtil.DeferClose(file)

	fmtUtil.Fprintf(file, "%s\n%s \"%s\"\n%s", ScriptHeader, script, strings.Join(args, "\" \""), ScriptFooter)
}
