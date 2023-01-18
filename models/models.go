package models

import (
	"strings"
)

type Config struct {
	Me        string
	Users     map[string]string
	ApiToken  string
	ApiUrl    string
	ProjectId string
}

type Flags struct {
	Env           string
	OnlyMine      bool
	NeedMyApprove bool
}

type Author struct {
	Username string `json:"username"`
}

type MergeRequests []struct {
	ID        int      `json:"id"`
	Iid       int      `json:"iid"`
	ProjectID int      `json:"project_id"`
	Title     string   `json:"title"`
	Author    Author   `json:"author"`
	Labels    []string `json:"labels"`
	WebURL    string   `json:"web_url"`
}

type Approval struct {
	ApprovedBy []struct {
		User struct {
			Id        int    `json:"id"`
			Username  string `json:"username"`
			Name      string `json:"name"`
			State     string `json:"state"`
			AvatarUrl string `json:"avatar_url"`
			WebUrl    string `json:"web_url"`
		} `json:"user"`
	} `json:"approved_by"`
}

type Discussions []struct {
	Id    string `json:"id"`
	Notes []struct {
		NoteableId int  `json:"noteable_id"`
		Resolvable bool `json:"resolvable"`
		Position   struct {
			HeadSha string `json:"head_sha"`
		} `json:"position,omitempty"`
		Resolved bool `json:"resolved,omitempty"`
	} `json:"notes"`
}

type Pipelines []struct {
	ID int `json:"id"`
}

type PipelineSummary struct {
	Total struct {
		Time       float64     `json:"time"`
		Count      int         `json:"count"`
		Success    int         `json:"success"`
		Failed     int         `json:"failed"`
		Skipped    int         `json:"skipped"`
		Error      int         `json:"error"`
		SuiteError interface{} `json:"suite_error"`
	} `json:"total"`
	TestSuites []struct {
		Name         string      `json:"name"`
		TotalTime    float64     `json:"total_time"`
		TotalCount   int         `json:"total_count"`
		SuccessCount int         `json:"success_count"`
		FailedCount  int         `json:"failed_count"`
		SkippedCount int         `json:"skipped_count"`
		ErrorCount   int         `json:"error_count"`
		BuildIds     []int       `json:"build_ids"`
		SuiteError   interface{} `json:"suite_error"`
	} `json:"test_suites"`
}

type NeedApprovals map[string]string

func (nA NeedApprovals) String() string {
	var sb strings.Builder
	delim := ""
	for _, user := range nA {
		sb.WriteString(delim + user)
		delim = "\n"
	}
	return sb.String()
}

type HasApprovals map[string]string

func (hA HasApprovals) String() string {
	var sb strings.Builder
	delim := ""
	for _, user := range hA {
		sb.WriteString(delim + user)
		delim = "\n"
	}
	return sb.String()
}

type NeedDiscs map[string]string

type Result struct {
	Iid           int
	Title         string
	WebURL        string
	Author        Author
	HasApprovals  HasApprovals
	NeedApprovals NeedApprovals
	NeedDiscs     NeedDiscs
	Pipelines     map[int]string
	PipelineFail  bool
}
