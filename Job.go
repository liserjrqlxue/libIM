package libIM

import (
	"strings"
)

type Job struct {
	ComputingFlag string                             `json:"computingFlag"`
	Mem           int                                `json:"mem"`
	Sh            string                             `json:"sh"`
	Step          *Step                              `json:"-"`
	Id            string                             `json:"-"`
	waitChan      map[string]map[string]*chan string // Step.Name->Job.Id->chan
	SubmitArgs    []string
}

func NewJob(sh string, mem int) Job {
	return Job{
		Mem:           mem,
		ComputingFlag: "cpu",
		Sh:            sh,
	}
}

func (job *Job) CreateWaitChan() {
	job.waitChan = make(map[string]map[string]*chan string)
	var step = job.Step
	for _, nextStep := range step.nextStep {
		var router = step.stepType + "->" + nextStep.stepType
		var chanMap = make(map[string]*chan string)
		switch router {
		case "batch->batch", "sample->batch", "lane->batch":
			var ch = make(chan string, 1)
			chanMap[nextStep.Name] = &ch
			job.waitChan[nextStep.Name] = chanMap
		case "batch->sample", "batch->lane":
			for jid := range nextStep.jobMap {
				var ch = make(chan string, 1)
				chanMap[jid] = &ch
			}
			job.waitChan[nextStep.Name] = chanMap
		case "sample->sample", "lane->lane":
			var ch = make(chan string, 1)
			chanMap[job.Id] = &ch
			job.waitChan[nextStep.Name] = chanMap
		case "sample->lane":
			for jid := range nextStep.jobMap {
				if strings.Split(jid, ":")[0] == job.Id {
					var ch = make(chan string, 1)
					chanMap[jid] = &ch
				}
			}
			job.waitChan[nextStep.Name] = chanMap
		case "lane->sample":
			var ch = make(chan string, 1)
			chanMap[strings.Split(job.Id, ":")[0]] = &ch
			job.waitChan[nextStep.Name] = chanMap
		}
	}
}

func (job *Job) WaitPriorChan() (jids []string) {
	var jidMap = make(map[string]bool)
	var step = job.Step
	for _, priorStep := range step.priorStep {
		var router = priorStep.stepType + "->" + step.stepType
		switch router {
		case "batch->batch", "batch->sample", "batch->lane":
			var priorJob = priorStep.jobMap[priorStep.Name]
			var ch = priorJob.waitChan[step.Name][job.Id]
			var jid = <-*ch
			if jid != "" {
				jidMap[jid] = true
			}
		case "sample->batch", "lane->batch":
			for _, priorJob := range priorStep.jobMap {
				var ch = priorJob.waitChan[step.Name][job.Id]
				var jid = <-*ch
				if jid != "" {
					jidMap[jid] = true
				}

			}
		case "sample->sample", "lane->lane":
			var priorJob = priorStep.jobMap[job.Id]
			var ch = priorJob.waitChan[step.Name][job.Id]
			var jid = <-*ch
			if jid != "" {
				jidMap[jid] = true
			}
		case "sample->lane":
			var priorJob = priorStep.jobMap[strings.Split(job.Id, ":")[0]]
			var ch = priorJob.waitChan[step.Name][job.Id]
			var jid = <-*ch
			if jid != "" {
				jidMap[jid] = true
			}
		case "lane->sample":
			for id, priorJob := range priorStep.jobMap {
				if strings.Split(id, ":")[0] == job.Id {
					var ch = priorJob.waitChan[step.Name][job.Id]
					var jid = <-*ch
					if jid != "" {
						jidMap[jid] = true
					}
				}
			}
		}
	}
	for jid := range jidMap {
		jids = append(jids, jid)
	}
	return
}

func (job *Job) Done(jid string) {
	for _, chs := range job.waitChan {
		for _, ch := range chs {
			*ch <- jid
		}
	}
}
