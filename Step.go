package libIM

import (
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/stringsUtil"
)

type Step struct {
	Name          string          `json:"name"`
	First         int             `json:"first"`
	StepFlag      int             `json:"stepFlag"`
	ComputingFlag string          `json:"computingFlag"`
	Memory        int             `json:"memory"`
	Threads       int             `json:"threads"`
	Timeout       int             `json:"timeout"`
	ModuleIndex   int             `json:"moduleIndex"`
	PriorStep     []string        `json:"priorStep"`
	NextStep      []string        `json:"nextStep"`
	JobSh         []*Job          `json:"jobSh"`
	jobMap        map[string]*Job `json:"-"`
	prior         string          `json:"-"` // ignore to json
	next          string          `json:"-"` // ignore to json
	stepType      string
	stepArgs      []string
}

func NewStep(item map[string]string) Step {
	return Step{
		Name:          item["name"],
		ComputingFlag: "cpu",
		Threads:       stringsUtil.Atoi(item["thread"]),
		Memory:        stringsUtil.Atoi(item["mem"]),
		// keep priorStep and nextStep [] instead null
		PriorStep: []string{},
		NextStep:  []string{},
		prior:     item["prior"],
		next:      item["next"],
		stepType:  item["type"],
		stepArgs:  strings.Split(item["args"], ","),
		jobMap:    make(map[string]*Job),
	}
}

func LinkSteps(stepMap map[string]*Step) {
	for stepName, step := range stepMap {
		for _, prior := range strings.Split(step.prior, ",") {
			priorStep, ok := stepMap[prior]
			if ok {
				step.PriorStep = append(step.PriorStep, prior)
				priorStep.NextStep = append(priorStep.NextStep, stepName)
			}
		}
	}
}

func (step *Step) CreateJobs(
	FamilyMap map[string]FamilyInfo, infoMap map[string]Info, trioInfo map[string]bool, workDir, pipeline string) (c int) {
	// script format: pipeline/script/stepName.sh
	var script = filepath.Join(pipeline, "script", strings.Join([]string{step.Name, "sh"}, "."))

	switch step.stepType {
	case "lane":
		return step.CreateLaneJobs(infoMap, workDir, pipeline, script)
	case "sample":
		return step.CreateSampleJobs(infoMap, workDir, pipeline, script)
	case "single":
		return step.CreateSingleJobs(infoMap, trioInfo, workDir, pipeline, script)
	case "trio":
		return step.CreateTrioJobs(infoMap, FamilyMap, workDir, pipeline, script)
	case "batch":
		return step.CreateBatchJob(workDir, pipeline, script)
	default:
		return
	}
}

func (step *Step) CreateLaneJobs(
	infoMap map[string]Info, workDir, pipeline, script string) (c int) {
	for sampleID, info := range infoMap {
		for _, lane := range info.LaneInfos {
			c++
			var job = step.CreateLaneJob(lane, workDir, pipeline, script, sampleID)
			step.jobMap[sampleID+":"+lane.LaneName] = &job
			step.JobSh = append(step.JobSh, &job)
		}
	}
	return
}

func (step *Step) CreateLaneJob(lane LaneInfo, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir, sampleID, "shell",
			strings.Join([]string{step.Name, lane.LaneName, "sh"}, "."),
		),
		step.Memory,
	)

	var args = []string{workDir, pipeline, sampleID}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneName":
			args = append(args, lane.LaneName)
		case "fq1":
			args = append(args, lane.Fq1)
		case "fq2":
			args = append(args, lane.Fq2)
		}
	}
	CreateShell(job.Sh, script, args...)
	return job
}

func (step *Step) CreateSingleJobs(
	infoMap map[string]Info, trioInfo map[string]bool, workDir, pipeline, script string) (c int) {
	for sampleID, info := range infoMap {
		if trioInfo[info.ProductCode] {
			continue
		}
		c++
		var job = step.CreateSampleJob(info, workDir, pipeline, script, sampleID)
		step.jobMap[sampleID] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateSampleJobs(
	infoMap map[string]Info, workDir, pipeline, script string) (c int) {
	for sampleID, info := range infoMap {
		c++
		var job = step.CreateSampleJob(info, workDir, pipeline, script, sampleID)
		step.jobMap[sampleID] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateSampleJob(info Info, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir, sampleID, "shell",
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)

	var args = []string{workDir, pipeline, sampleID}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneName":
			for _, lane := range info.LaneInfos {
				args = append(args, lane.LaneName)
			}
		case "gender":
			args = append(args, info.Gender)
		case "HPO":
			args = append(args, info.HPO)
		case "StandardTag":
			args = append(args, info.StandardTag)
		case "product_code":
			args = append(args, info.ProductCode)
		case "QChistory":
			args = append(args, info.QChistory)
		case "chip_code":
			args = append(args, info.ChipCode)
		}
	}
	CreateShell(job.Sh, script, args...)
	return job
}

func (step *Step) CreateTrioJobs(
	infoMap map[string]Info, familyInfoMap map[string]FamilyInfo, workDir, pipeline, script string) (c int) {
	for probandID, familyInfo := range familyInfoMap {
		c++
		var job = step.CreateTrioJob(infoMap[probandID], familyInfo, workDir, pipeline, script, probandID)
		step.jobMap[probandID] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateTrioJob(info Info, familyInfo FamilyInfo, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir, sampleID, "shell",
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)

	var args = []string{workDir, pipeline}
	for _, arg := range step.stepArgs {
		switch arg {
		case "list":
			for _, relationShip := range Trio {
				args = append(args, familyInfo.FamilyMap[relationShip])
			}
		case "HPO":
			args = append(args, info.HPO)
		case "product_code":
			args = append(args, info.ProductCode)
		}
	}
	CreateShell(job.Sh, script, args...)
	return job
}

func (step *Step) CreateBatchJob(workDir, pipeline, script string) (c int) {
	var job = NewJob(
		filepath.Join(
			workDir, "shell",
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)

	var args = []string{workDir, pipeline}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneInput":
			args = append(args, LaneInput)
		}
	}
	CreateShell(job.Sh, script, args...)

	c++
	step.JobSh = append(step.JobSh, &job)
	step.jobMap[step.Name] = &job
	return
}
