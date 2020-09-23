package slackstructs

type SlackAWSCredCheckerStyle struct {
	Blocks []SlackAWSCredCheckerBlock `json:"blocks"`
}

type SlackAWSCredCheckerBlock struct {
	HeaderType     string                           `json:"type"`
	HeaderText     SlackAWSCredCheckerHeaderText    `json:"text"`
	SectionType    string                           `json:"type"`
	SectionText    SlackAWSCredCheckerSectionText   `json:"text"`
	ActionType     string                           `json:"type"`
	ActionElements []SlackAWSCredCheckerActionBlock `json:"elements"`
}

type SlackAWSCredCheckerActionBlock struct {
	Type        string                        `json:"type"`
	ElementText SlackAWSCredCheckerHeaderText `json:"text"`
	Value       string                        `json:"value"`
}

type SlackAWSCredCheckerHeaderText struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}

type SlackAWSCredCheckerSectionText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
