package libIM

import (
	"strings"

	"github.com/liserjrqlxue/goUtil/fmtUtil"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func CreateShell(fileName, script string, args ...string) {
	var file = osUtil.Create(fileName)
	defer simpleUtil.DeferClose(file)

	fmtUtil.Fprintf(file, "#!/bin/bash\nsh %s %s\n", script, strings.Join(args, " "))
}
