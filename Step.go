package libIM

import (
	"path/filepath"
	"strings"

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
	JobSh         []Job    `json:"jobSh"`
	Prior, Next   string
	stepType      string
	stepArgs      []string
}

func NewStep(item map[string]string) (step Step) {
	step = Step{}
	step.Name = item["name"]
	step.ComputingFlag = "cpu"
	step.Threads = stringsUtil.Atoi(item["thread"])
	step.Memory = stringsUtil.Atoi(item["mem"])
	step.PriorStep = []string{}
	step.NextStep = []string{}
	step.Prior = item["prior"]
	step.Next = item["next"]
	step.stepType = item["type"]
	step.stepArgs = strings.Split(item["args"], ",")
	return
}

func (step *Step) CreateJobs(
	FamilyMap map[string]FamilyInfo, infoMap map[string]Info, trioInfo map[string]bool, workDir, pipeline string) {
	// script format: pipeline/script/stepName.sh
	var script = filepath.Join(pipeline, "script", step.Name, ".sh")
	var jobs []Job

	switch step.stepType {
	case "lane":
		jobs = step.CreateLaneJobs(infoMap, workDir, pipeline, script)
	case "sample":
		jobs = step.CreateSampleJobs(infoMap, workDir, pipeline, script)
	case "single":
		jobs = step.CreateSingleJobs(infoMap, trioInfo, workDir, pipeline, script)
	case "trio":
		jobs = step.CreateTrioJobs(infoMap, FamilyMap, workDir, pipeline, script)
	case "batch":
		jobs = step.CreateBatchJob(workDir, pipeline, script)
	}

	step.JobSh = jobs
}

func (step *Step) CreateLaneJobs(infoMap map[string]Info, workDir, pipeline, script string) (jobs []Job) {
	for sampleID, info := range infoMap {
		for _, lane := range info.LaneInfos {
			jobs = append(jobs, step.CreateLaneJob(lane, workDir, pipeline, script, sampleID))
		}
	}
	return
}

func (step *Step) CreateLaneJob(lane LaneInfo, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(step.Memory)
	var args = []string{workDir, pipeline, sampleID}

	job.Sh = filepath.Join(
		workDir, sampleID, "shell",
		strings.Join([]string{step.Name, lane.LaneName, "sh"}, "."),
	)

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
	infoMap map[string]Info, trioInfo map[string]bool, workDir, pipeline, script string) (jobs []Job) {
	for sampleID, info := range infoMap {
		if trioInfo[info.ProductCode] {
			continue
		}
		jobs = append(jobs, step.CreateSampleJob(info, workDir, pipeline, script, sampleID))
	}
	return
}

func (step *Step) CreateSampleJobs(infoMap map[string]Info, workDir, pipeline, script string) (jobs []Job) {
	for sampleID, info := range infoMap {
		jobs = append(jobs, step.CreateSampleJob(info, workDir, pipeline, script, sampleID))
	}
	return
}

func (step *Step) CreateSampleJob(info Info, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(step.Memory)
	job.Sh = filepath.Join(
		workDir, sampleID, "shell",
		strings.Join([]string{step.Name, "sh"}, "."),
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
	infoMap map[string]Info, familyInfoMap map[string]FamilyInfo, workDir, pipeline, script string) (jobs []Job) {
	for probandID, familyInfo := range familyInfoMap {
		jobs = append(jobs, step.CreateTrioJob(infoMap[probandID], familyInfo, workDir, pipeline, script, probandID))
	}
	return
}

func (step *Step) CreateTrioJob(info Info, familyInfo FamilyInfo, workDir, pipeline, script, sampleID string) Job {
	var job = NewJob(step.Memory)
	job.Sh = filepath.Join(
		workDir, sampleID, "shell",
		strings.Join([]string{step.Name, "sh"}, "."),
	)

	var args = []string{workDir, pipeline}
	for _, arg := range step.stepArgs {
		switch arg {
		case "list":
			for _, relationShip := range []string{"proband", "father", "mother"} {
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

func (step *Step) CreateBatchJob(workDir, pipeline, script string) (jobs []Job) {
	var job = NewJob(step.Memory)
	job.Sh = filepath.Join(
		workDir, "shell",
		strings.Join([]string{step.Name, "sh"}, "."),
	)

	var args = []string{workDir, pipeline}
	for _, arg := range step.stepArgs {
		switch arg {
		case "laneInput":
			args = append(args, LaneInput)
		}
	}
	CreateShell(job.Sh, script, args...)
	return
}
