package github

import (
	"fmt"
	"log"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/metric"
)

const meas = "github_webhooks"

type Event interface {
	NewMetric() telex.Metric
}

type Repository struct {
	Repository string `json:"full_name"`
	Private    bool   `json:"private"`
	Stars      int    `json:"stargazers_count"`
	Forks      int    `json:"forks_count"`
	Issues     int    `json:"open_issues_count"`
}

type Sender struct {
	User  string `json:"login"`
	Admin bool   `json:"site_admin"`
}

type CommitComment struct {
	Commit string `json:"commit_id"`
	Body   string `json:"body"`
}

type Deployment struct {
	Commit      string `json:"sha"`
	Task        string `json:"task"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

type Page struct {
	Name   string `json:"page_name"`
	Title  string `json:"title"`
	Action string `json:"action"`
}

type Issue struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Comments int    `json:"comments"`
}

type IssueComment struct {
	Body string `json:"body"`
}

type Team struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Number       int    `json:"number"`
	State        string `json:"state"`
	Title        string `json:"title"`
	Comments     int    `json:"comments"`
	Commits      int    `json:"commits"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ChangedFiles int    `json:"changed_files"`
}

type PullRequestReviewComment struct {
	File    string `json:"path"`
	Comment string `json:"body"`
}

type Release struct {
	TagName string `json:"tag_name"`
}

type DeploymentStatus struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

type CommitCommentEvent struct {
	Comment    CommitComment `json:"comment"`
	Repository Repository    `json:"repository"`
	Sender     Sender        `json:"sender"`
}

func (s CommitCommentEvent) NewMetric() telex.Metric {
	event := "commit_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"commit":  s.Comment.Commit,
		"comment": s.Comment.Body,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type CreateEvent struct {
	Ref        string     `json:"ref"`
	RefType    string     `json:"ref_type"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s CreateEvent) NewMetric() telex.Metric {
	event := "create"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"ref":     s.Ref,
		"refType": s.RefType,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type DeleteEvent struct {
	Ref        string     `json:"ref"`
	RefType    string     `json:"ref_type"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s DeleteEvent) NewMetric() telex.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"ref":     s.Ref,
		"refType": s.RefType,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type DeploymentEvent struct {
	Deployment Deployment `json:"deployment"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s DeploymentEvent) NewMetric() telex.Metric {
	event := "deployment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":       s.Repository.Stars,
		"forks":       s.Repository.Forks,
		"issues":      s.Repository.Issues,
		"commit":      s.Deployment.Commit,
		"task":        s.Deployment.Task,
		"environment": s.Deployment.Environment,
		"description": s.Deployment.Description,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type DeploymentStatusEvent struct {
	Deployment       Deployment       `json:"deployment"`
	DeploymentStatus DeploymentStatus `json:"deployment_status"`
	Repository       Repository       `json:"repository"`
	Sender           Sender           `json:"sender"`
}

func (s DeploymentStatusEvent) NewMetric() telex.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":          s.Repository.Stars,
		"forks":          s.Repository.Forks,
		"issues":         s.Repository.Issues,
		"commit":         s.Deployment.Commit,
		"task":           s.Deployment.Task,
		"environment":    s.Deployment.Environment,
		"description":    s.Deployment.Description,
		"depState":       s.DeploymentStatus.State,
		"depDescription": s.DeploymentStatus.Description,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type ForkEvent struct {
	Forkee     Repository `json:"forkee"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s ForkEvent) NewMetric() telex.Metric {
	event := "fork"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"fork":   s.Forkee.Repository,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type GollumEvent struct {
	Pages      []Page     `json:"pages"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

// REVIEW: Going to be lazy and not deal with the pages.
func (s GollumEvent) NewMetric() telex.Metric {
	event := "gollum"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type IssueCommentEvent struct {
	Issue      Issue        `json:"issue"`
	Comment    IssueComment `json:"comment"`
	Repository Repository   `json:"repository"`
	Sender     Sender       `json:"sender"`
}

func (s IssueCommentEvent) NewMetric() telex.Metric {
	event := "issue_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
		"issue":      fmt.Sprintf("%v", s.Issue.Number),
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"title":    s.Issue.Title,
		"comments": s.Issue.Comments,
		"body":     s.Comment.Body,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type IssuesEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s IssuesEvent) NewMetric() telex.Metric {
	event := "issue"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
		"issue":      fmt.Sprintf("%v", s.Issue.Number),
		"action":     s.Action,
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"title":    s.Issue.Title,
		"comments": s.Issue.Comments,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type MemberEvent struct {
	Member     Sender     `json:"member"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s MemberEvent) NewMetric() telex.Metric {
	event := "member"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":           s.Repository.Stars,
		"forks":           s.Repository.Forks,
		"issues":          s.Repository.Issues,
		"newMember":       s.Member.User,
		"newMemberStatus": s.Member.Admin,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type MembershipEvent struct {
	Action string `json:"action"`
	Member Sender `json:"member"`
	Sender Sender `json:"sender"`
	Team   Team   `json:"team"`
}

func (s MembershipEvent) NewMetric() telex.Metric {
	event := "membership"
	t := map[string]string{
		"event":  event,
		"user":   s.Sender.User,
		"admin":  fmt.Sprintf("%v", s.Sender.Admin),
		"action": s.Action,
	}
	f := map[string]interface{}{
		"newMember":       s.Member.User,
		"newMemberStatus": s.Member.Admin,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type PageBuildEvent struct {
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s PageBuildEvent) NewMetric() telex.Metric {
	event := "page_build"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type PublicEvent struct {
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s PublicEvent) NewMetric() telex.Metric {
	event := "public"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      Sender      `json:"sender"`
}

func (s PullRequestEvent) NewMetric() telex.Metric {
	event := "pull_request"
	t := map[string]string{
		"event":      event,
		"action":     s.Action,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
		"prNumber":   fmt.Sprintf("%v", s.PullRequest.Number),
	}
	f := map[string]interface{}{
		"stars":        s.Repository.Stars,
		"forks":        s.Repository.Forks,
		"issues":       s.Repository.Issues,
		"state":        s.PullRequest.State,
		"title":        s.PullRequest.Title,
		"comments":     s.PullRequest.Comments,
		"commits":      s.PullRequest.Commits,
		"additions":    s.PullRequest.Additions,
		"deletions":    s.PullRequest.Deletions,
		"changedFiles": s.PullRequest.ChangedFiles,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type PullRequestReviewCommentEvent struct {
	Comment     PullRequestReviewComment `json:"comment"`
	PullRequest PullRequest              `json:"pull_request"`
	Repository  Repository               `json:"repository"`
	Sender      Sender                   `json:"sender"`
}

func (s PullRequestReviewCommentEvent) NewMetric() telex.Metric {
	event := "pull_request_review_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
		"prNumber":   fmt.Sprintf("%v", s.PullRequest.Number),
	}
	f := map[string]interface{}{
		"stars":        s.Repository.Stars,
		"forks":        s.Repository.Forks,
		"issues":       s.Repository.Issues,
		"state":        s.PullRequest.State,
		"title":        s.PullRequest.Title,
		"comments":     s.PullRequest.Comments,
		"commits":      s.PullRequest.Commits,
		"additions":    s.PullRequest.Additions,
		"deletions":    s.PullRequest.Deletions,
		"changedFiles": s.PullRequest.ChangedFiles,
		"commentFile":  s.Comment.File,
		"comment":      s.Comment.Comment,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type PushEvent struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s PushEvent) NewMetric() telex.Metric {
	event := "push"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"ref":    s.Ref,
		"before": s.Before,
		"after":  s.After,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type ReleaseEvent struct {
	Release    Release    `json:"release"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s ReleaseEvent) NewMetric() telex.Metric {
	event := "release"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"tagName": s.Release.TagName,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type RepositoryEvent struct {
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s RepositoryEvent) NewMetric() telex.Metric {
	event := "repository"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type StatusEvent struct {
	Commit     string     `json:"sha"`
	State      string     `json:"state"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s StatusEvent) NewMetric() telex.Metric {
	event := "status"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"commit": s.Commit,
		"state":  s.State,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type TeamAddEvent struct {
	Team       Team       `json:"team"`
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s TeamAddEvent) NewMetric() telex.Metric {
	event := "team_add"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"teamName": s.Team.Name,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}

type WatchEvent struct {
	Repository Repository `json:"repository"`
	Sender     Sender     `json:"sender"`
}

func (s WatchEvent) NewMetric() telex.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    fmt.Sprintf("%v", s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      fmt.Sprintf("%v", s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", event)
	}
	return m
}
