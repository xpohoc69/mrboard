package requesters

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xpohoc69/mrboard/models"
	"io"
	"net/http"
	"time"
)

type GitlabRequester struct {
	httpClient http.Client
	config     *models.Config
}

func NewRequester(config *models.Config) *GitlabRequester {
	return &GitlabRequester{
		httpClient: http.Client{Timeout: 5 * time.Second},
		config:     config,
	}
}

func (r GitlabRequester) doGetRequest(url string) (response []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}
	req.Header = http.Header{
		"PRIVATE-TOKEN": {r.config.ApiToken},
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		return response, errors.New("empty body")
	}

	response, err = io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	return response, err
}

func (r GitlabRequester) GetMergeRequests() (mergeRequests models.MergeRequests, err error) {
	url := r.config.ApiUrl + "/merge_requests?scope=all&state=opened"

	response, err := r.doGetRequest(url)
	if err != nil {
		return mergeRequests, err
	}
	err = json.Unmarshal(response, &mergeRequests)

	return mergeRequests, err
}

func (r GitlabRequester) GetApprovals(mergeRequestIid int) (approval models.Approval, err error) {
	url := fmt.Sprintf("%v/projects/%v/merge_requests/%v/approvals", r.config.ApiUrl, r.config.ProjectId, mergeRequestIid)
	response, err := r.doGetRequest(url)
	if err != nil {
		return approval, err
	}
	err = json.Unmarshal(response, &approval)

	return approval, err
}

func (r GitlabRequester) GetDiscussions(mergeRequestIid int) (discussions models.Discussions, err error) {
	url := fmt.Sprintf("%v/projects/%v/merge_requests/%v/discussions?per_page=100", r.config.ApiUrl, r.config.ProjectId, mergeRequestIid)
	response, err := r.doGetRequest(url)
	if err != nil {
		return discussions, err
	}
	err = json.Unmarshal(response, &discussions)

	return discussions, err
}

func (r GitlabRequester) GetPipelines(mergeRequestIid int) (pipelines models.Pipelines, err error) {
	url := fmt.Sprintf("%v/projects/%v/merge_requests/%v/pipelines", r.config.ApiUrl, r.config.ProjectId, mergeRequestIid)
	response, err := r.doGetRequest(url)
	if err != nil {
		return pipelines, err
	}
	err = json.Unmarshal(response, &pipelines)

	return pipelines, err
}

func (r GitlabRequester) GetPipelineReport(id int) (pipelineSummary models.PipelineSummary, err error) {
	url := fmt.Sprintf("%v/projects/%v/pipelines/%v/test_report_summary", r.config.ApiUrl, r.config.ProjectId, id)
	response, err := r.doGetRequest(url)
	if err != nil {
		return pipelineSummary, err
	}
	err = json.Unmarshal(response, &pipelineSummary)

	return pipelineSummary, err
}
