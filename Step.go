package libIM

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/stringsUtil"
)

type Step struct {
	Name          string   `json:"name"`
	First         int      `json:"first"`
	StepFlag      int      `json:"stepFlag"`
	ComputingFlag string   `json:"computingFlag"`
	Memory        int      `json:"memory"`
	Threads       int      `json:"threads"`
	Timeout       int      `json:"timeout"`
	ModuleIndex   int      `json:"moduleIndex"`
	PriorStep     []string `json:"priorStep"`
	NextStep      []string `json:"nextStep"`
	JobSh         []*Job   `json:"jobSh"`
	script        string
	jobMap        map[string]*Job
	prior         string
	next          string
	stepType      string
	priorStep     []*Step
	nextStep      []*Step
	stepArgs      []string
	submitArgs    []string
}

func NewStep(item map[string]string) Step {
	return Step{
		Name:          item["name"],
		ComputingFlag: "cpu",
		Threads:       stringsUtil.Atoi(item["thread"]),
		Memory:        stringsUtil.Atoi(item["mem"]),
		// keep priorStep and nextStep [] instead null
		PriorStep:  []string{},
		NextStep:   []string{},
		prior:      item["prior"],
		next:       item["next"],
		script:     item["script"],
		stepType:   item["type"],
		stepArgs:   strings.Split(item["args"], ","),
		submitArgs: strings.Split(item["submitArgs"], " "),
		jobMap:     make(map[string]*Job),
	}
}

func LinkSteps(stepMap map[string]*Step) {
	for stepName, step := range stepMap {
		for _, prior := range strings.Split(step.prior, ",") {
			priorStep, ok := stepMap[prior]
			if ok {
				step.PriorStep = append(step.PriorStep, prior)
				step.priorStep = append(step.priorStep, priorStep)
				priorStep.NextStep = append(priorStep.NextStep, stepName)
				priorStep.nextStep = append(priorStep.nextStep, step)
			}
		}
	}
	for _, step := range stepMap {
		for _, job := range step.jobMap {
			job.CreateWaitChan()
		}
	}
}

func (step *Step) CreateJobs(
	FamilyMap map[string]FamilyInfo, infoMap map[string]*Info, trioInfo map[string]bool, workDir, pipeline string) (c int) {
	// script format: pipeline/script/stepName.sh
	if step.script == "" {
		step.script = filepath.Join(pipeline, "script", strings.Join([]string{step.Name, "sh"}, "."))
	} else if !osUtil.FileExists(step.script) {
		step.script = filepath.Join(pipeline, "script", step.script)
	}
	if !osUtil.FileExists(step.script) {
		log.Fatalf("can not find [%s] script:[%s]\n", step.Name, step.script)
	}

	switch step.stepType {
	case "lane":
		return step.CreateLaneJobs(infoMap, workDir, pipeline)
	case "sample":
		return step.CreateSampleJobs(infoMap, workDir, pipeline)
	case "single":
		return step.CreateSingleJobs(infoMap, trioInfo, workDir, pipeline)
	case "trio":
		return step.CreateTrioJobs(infoMap, FamilyMap, workDir, pipeline)
	case "batch":
		return step.CreateBatchJob(workDir, pipeline)
	default:
		return
	}
}

func (step *Step) CreateLaneJobs(
	infoMap map[string]*Info, workDir, pipeline string) (c int) {
	for sampleID, info := range infoMap {
		for _, lane := range info.LaneInfos {
			c++
			var job = step.CreateLaneJob(lane, workDir, pipeline, sampleID)
			job.SubmitArgs = step.submitArgs
			step.jobMap[job.Id] = &job
			step.JobSh = append(step.JobSh, &job)
		}
	}
	return
}

func (step *Step) CreateLaneJob(lane LaneInfo, workDir, pipeline, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir,
			"shell",
			sampleID,
			strings.Join([]string{step.Name, lane.LaneName, "sh"}, "."),
		),
		step.Memory,
	)
	job.Step = step
	job.Id = sampleID + ":" + lane.LaneName
	job.SubmitArgs = step.submitArgs

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
	CreateShell(job.Sh, step.script, args...)
	return job
}

func (step *Step) CreateSingleJobs(
	infoMap map[string]*Info, trioInfo map[string]bool, workDir, pipeline string) (c int) {
	for sampleID, info := range infoMap {
		if trioInfo[info.ProductCode] {
			continue
		}
		c++
		var job = step.CreateSampleJob(info, workDir, pipeline, sampleID)
		job.SubmitArgs = step.submitArgs
		step.jobMap[job.Id] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateSampleJobs(
	infoMap map[string]*Info, workDir, pipeline string) (c int) {
	for sampleID, info := range infoMap {
		c++
		var job = step.CreateSampleJob(info, workDir, pipeline, sampleID)
		job.SubmitArgs = step.submitArgs
		step.jobMap[job.Id] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateSampleJob(info *Info, workDir, pipeline, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir,
			"shell",
			sampleID,
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)
	job.Step = step
	job.Id = sampleID
	job.SubmitArgs = step.submitArgs

	var args = []string{workDir, pipeline, sampleID}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneName":
			for _, lane := range info.LaneInfos {
				args = append(args, lane.LaneName)
			}
		case "fq1":
			for _, lane := range info.LaneInfos {
				args = append(args, lane.Fq1)
			}
		case "fq2":
			for _, lane := range info.LaneInfos {
				args = append(args, lane.Fq2)
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
		default:
			args = append(args, info.Raw[arg])
		}
	}
	CreateShell(job.Sh, step.script, args...)
	return job
}

func (step *Step) CreateTrioJobs(
	infoMap map[string]*Info, familyInfoMap map[string]FamilyInfo, workDir, pipeline string) (c int) {
	for probandID, familyInfo := range familyInfoMap {
		c++
		var job = step.CreateTrioJob(infoMap[probandID], familyInfo, workDir, pipeline, probandID)
		job.SubmitArgs = step.submitArgs
		step.jobMap[job.Id] = &job
		step.JobSh = append(step.JobSh, &job)
	}
	return
}

func (step *Step) CreateTrioJob(info *Info, familyInfo FamilyInfo, workDir, pipeline, sampleID string) Job {
	var job = NewJob(
		filepath.Join(
			workDir,
			"shell",
			sampleID,
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)
	job.Step = step
	job.Id = sampleID
	job.SubmitArgs = step.submitArgs

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
	CreateShell(job.Sh, step.script, args...)
	return job
}

func (step *Step) CreateBatchJob(workDir, pipeline string) (c int) {
	var job = NewJob(
		filepath.Join(
			workDir,
			"shell",
			"batch",
			strings.Join([]string{step.Name, "sh"}, "."),
		),
		step.Memory,
	)
	job.Step = step
	job.Id = step.Name
	job.SubmitArgs = step.submitArgs

	var args = []string{workDir, pipeline}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneInput":
			args = append(args, LaneInput)
		}
	}
	CreateShell(job.Sh, step.script, args...)

	c++
	step.jobMap[job.Id] = &job
	step.JobSh = append(step.JobSh, &job)
	return
}
