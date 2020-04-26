package libIM

type LaneInfo struct {
	LaneName string `json:"lane_code"`
	Fq1      string `json:"fastq1"`
	Fq2      string `json:"fastq2"`
}
type Info struct {
	SampleID     string
	ChipCode     string
	Type         string
	Gender       string
	ProductCode  string
	ProductType  string
	ProbandID    string
	RelationShip string
	HPO          string
	StandardTag  string
	StandardQC   string
	QChistory    string
	LaneInfos    []LaneInfo
	FamilyInfo   map[string][]string
}

type FamilyInfo struct {
	ProbandID string
	FamilyMap map[string]string
}

func NewInfo(item map[string]string) (info Info) {
	info = Info{}
	info.SampleID = item["main_sample_num"]
	info.ChipCode = item["chip_code"]
	info.Gender = item["gender"]
	info.ProductCode = item["product_code"]
	info.ProductType = item["probuctType"]
	info.ProbandID = item["proband_number"]
	info.HPO = item["HPO"]
	info.StandardTag = item["isStandardSample"]
	info.StandardQC = item["StandardQC"]
	info.RelationShip = item["relationship"]
	info.QChistory = item["QChistory"]
	return
}
