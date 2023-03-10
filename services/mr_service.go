package services

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/gookit/goutil/arrutil"
	"github.com/gookit/goutil/maputil"
	"github.com/olekukonko/tablewriter"
	"github.com/xpohoc69/mrboard/models"
	"github.com/xpohoc69/mrboard/requesters"
)

const progressSteps = 3

type MrService struct {
	mu        sync.RWMutex
	wg        sync.WaitGroup
	config    *models.Config
	requester *requesters.GitlabRequester
	flags     *models.Flags
}

func NewMrService(config *models.Config, requester *requesters.GitlabRequester, flags *models.Flags) *MrService {
	return &MrService{
		config:    config,
		requester: requester,
		flags:     flags,
	}
}

func (s *MrService) PrepareResult() map[int]*models.Result {
	result := make(map[int]*models.Result, 8)
	progressBar := pb.Simple.Start(progressSteps)
	progressBar.Increment()
	mergeRequests, err := s.requester.GetMergeRequests()
	if err != nil {
		log.Fatal(err)
	}
	mergeRequests = s.filterMergeRequests(mergeRequests)

	for _, mR := range mergeRequests {
		result[mR.Iid] = &models.Result{
			Iid:           mR.Iid,
			ProjectID:     mR.ProjectID,
			Title:         mR.Title,
			WebURL:        mR.WebURL,
			Author:        mR.Author,
			NeedApprovals: make(map[string]string, 5),
			HasApprovals:  make(map[string]string, 5),
			NeedDiscs:     make(map[string]string, 5),
		}
	}

	progressBar.Increment()
	result = s.getApprovals(result)
	result = s.getDiscussions(result)
	pipelines := s.getPipelines(result)
	s.wg.Wait()
	progressBar.Increment()
	result = s.getFailedPipelines(result, pipelines)
	progressBar.Finish()

	return result
}

func (s *MrService) PrintTable(result map[int]*models.Result) {
	headers := []string{"Task", "MR", "Author", "Given approvals", "Need approvals", "Open discussions count", "Is last pipeline failed"}
	if s.config.TaskIdRegex == "" {
		headers = headers[1:]
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader(headers)

	keys := make([]int, 0, len(result))
	for iid := range result {
		keys = append(keys, iid)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	for _, iid := range keys {
		item := result[iid]

		if s.flags.OnlyMine {
			if item.Author.Username != s.config.Me {
				continue
			}
		}
		if s.flags.NeedMyApprove {
			if !s.flags.OnlyMine && item.Author.Username == s.config.Me {
				continue
			}
			if _, ok := item.HasApprovals[s.config.Me]; ok {
				continue
			}
		}

		taskUrl := " "
		if s.config.TaskIdRegex != "" {
			re := regexp.MustCompile(s.config.TaskIdRegex)
			match := re.FindAllString(item.Title, -1)
			if len(match) > 0 {
				taskUrl = s.config.TaskUrl + match[0]
			}
		}

		hasFailedPipeline := "-"
		if item.PipelineFail {
			hasFailedPipeline = "+"
		}

		row := []string{
			taskUrl,
			fmt.Sprintf("%v (%v)", item.Title, item.WebURL),
			item.Author.Username,
			fmt.Sprint(item.HasApprovals),
			fmt.Sprint(item.NeedApprovals),
			strconv.Itoa(len(item.NeedDiscs)),
			hasFailedPipeline,
		}

		if taskUrl == "" {
			row = row[1:]
		}

		table.Append(row)
	}

	fmt.Print("\n")
	table.Render()
}

func (s *MrService) getApprovals(result map[int]*models.Result) map[int]*models.Result {
	s.mu.RLock()
	for _, item := range result {
		item := item
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			approval, err := s.requester.GetApprovals(item)
			if err != nil {
				log.Println(err)
				return
			}

			copyUsers := make(map[string]string)
			for index, element := range s.config.Users {
				copyUsers[index] = element
			}
			delete(copyUsers, item.Author.Username)

			if len(approval.ApprovedBy) != 0 {
				for _, approval := range approval.ApprovedBy {
					item.HasApprovals[approval.User.Username] = approval.User.Username
					delete(copyUsers, approval.User.Username)
				}
			}
			s.mu.Lock()
			item.NeedApprovals = copyUsers
			result[item.Iid] = item
			s.mu.Unlock()
		}()
	}
	s.mu.RUnlock()

	return result
}

func (s *MrService) getDiscussions(result map[int]*models.Result) map[int]*models.Result {
	s.mu.RLock()
	for _, item := range result {
		item := item
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			discussions, err := s.requester.GetDiscussions(item)
			if err != nil {
				log.Println(err)
				return
			}

			if len(discussions) == 0 {
				return
			}

			s.mu.Lock()
			for _, discussion := range discussions {
				for _, note := range discussion.Notes {
					if !note.Resolvable {
						continue
					}
					if note.Resolved {
						continue
					}
					position := strconv.Itoa(note.NoteableId)
					if len(note.Position.HeadSha) > 0 {
						position = note.Position.HeadSha
					}

					item.NeedDiscs[position] = position
				}
			}
			result[item.Iid] = item
			s.mu.Unlock()
		}()
	}
	s.mu.RUnlock()

	return result
}

func (s *MrService) getPipelines(result map[int]*models.Result) map[int]models.Pipelines {
	s.mu.RLock()
	pipelinesMap := make(map[int]models.Pipelines, 5)
	for _, item := range result {
		item := item
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			pipelines, err := s.requester.GetPipelines(item)
			if err != nil {
				log.Println(err)
				return
			}
			s.mu.Lock()
			pipelinesMap[item.Iid] = pipelines
			s.mu.Unlock()
		}()
	}
	s.mu.RUnlock()

	return pipelinesMap
}

func (s *MrService) getFailedPipelines(result map[int]*models.Result, pipelinesMap map[int]models.Pipelines) map[int]*models.Result {
	s.mu.RLock()
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for mergeRequestIid, pipelines := range pipelinesMap {
		mergeRequestIid := mergeRequestIid
		if len(pipelines) < 1 {
			continue
		}
		topPipeline := pipelines[0]
		if topPipeline.ID < 1 {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			pipelineSummary, err := s.requester.GetPipelineReport(topPipeline)
			if err != nil {
				log.Println(err)
				return
			}

			s.mu.Lock()
			if item, ok := result[mergeRequestIid]; ok {
				item.PipelineFail = false
				if pipelineSummary.Total.Failed > 0 || pipelineSummary.Total.Error > 0 {
					item.PipelineFail = true
				}
				result[mergeRequestIid] = item
			}
			s.mu.Unlock()
		}()
	}
	s.mu.RUnlock()

	return result
}

func (s *MrService) filterMergeRequests(mergeRequests models.MergeRequests) models.MergeRequests {
	filteredMergeRequests := models.MergeRequests{}
	for _, mR := range mergeRequests {
		if strings.Contains(mR.Title, "Draft:") ||
			strings.Contains(mR.WebURL, "services/front") ||
			arrutil.Contains(mR.Labels, "Revert of revert") ||
			!maputil.HasKey(s.config.Users, mR.Author.Username) {
			continue
		}
		filteredMergeRequests = append(filteredMergeRequests, mR)
	}

	return filteredMergeRequests
}
