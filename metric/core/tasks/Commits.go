package tasks

import (
	"core/models"
	"core/repositories"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

func getCommitBatch(endpoint string, page int, apiKeys []string) []interface{} {
	client := http.Client{}
	req, _ := http.NewRequest("GET", endpoint+"commits?per_page=100&page="+strconv.Itoa(page), nil)
	req.Header.Set("Authorization", "Bearer "+apiKeys[0])
	resp, err := client.Do(req)

	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil
	}

	var parsedBody []interface{}
	err = json.Unmarshal(body, &parsedBody)

	if err != nil {
		return nil
	}

	return parsedBody
}

func getCommitDetailed(endpoint string, apiKeys []string) (interface{}, error) {
	client := http.Client{}
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+apiKeys[0])
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("request failed")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var parsedBody interface{}
	err = json.Unmarshal(body, &parsedBody)

	if err != nil {
		return nil, err
	}

	return parsedBody, nil
}

func saveCommitsDetailed(commits []interface{}, files [][]byte, task models.Task, repo *repositories.SnapshotRepository) {
	var result []*models.Snapshot
	var filesResult []*models.Snapshot

	for _, commit := range commits {
		// Convert commit object to JSON
		data, err := json.Marshal(commit)
		if err != nil {
			continue
		}

		// Save commit id as a parameter in case we ever need to find it in the DB
		result = append(result, taskToSnapshot(task, string(data), "", []models.SnapshotParam{
			{
				Name:  "id",
				Value: commit.(map[string]interface{})["node_id"].(string),
			},
		}))
	}

	for _, data := range files {

		var file map[string]interface{}
		var err error
		json.Unmarshal(data, &file)

		if err != nil {
			continue
		}

		// Save commit sha as a parameter in case we ever need to find all commit files in the DB
		// Save the file id as a parameter in case we ever need to find it in the DB
		filesResult = append(filesResult, &models.Snapshot{
			Metric: "CommitFiles",
			Data:   string(data),
			Groups: task.Groups,
			Params: []models.SnapshotParam{
				{
					Name:  "id",
					Value: file["sha"].(string),
				},
				{
					Name:  "commit_sha",
					Value: file["commit_sha"].(string),
				},
			},
			Error:    "",
			IsPublic: false,
		})
	}

	if len(result) != 0 {
		repo.CreateInBatches(result)
	}

	if len(filesResult) != 0 {
		repo.CreateInBatches(filesResult)
	}
}

func CommitsMetric(task models.Task, repo *repositories.SnapshotRepository) {
	// Try to decode the JSON serialized task parameters
	var parsed []interface{}
	err := json.Unmarshal([]byte(task.Data), &parsed)
	if err != nil {
		repo.Create(taskToSnapshot(task, "", err.Error(), nil))
		return
	}

	// Find the endpoint parameter
	endpoint := getEndpoint(parsed)
	if endpoint == "" {
		repo.Create(taskToSnapshot(task, "", "no API endpoint", nil))
		return
	}

	// Find the API keys parameter
	apiKeys := getAPIKeys(parsed)
	if len(apiKeys) == 0 {
		repo.Create(taskToSnapshot(task, "", "no API keys", nil))
		return
	}

	var commits []interface{}
	var files [][]byte

	// GitHub API returns a maximum of 100 commits per request
	page := 1
	commitsBatch := getCommitBatch(endpoint, page, apiKeys)

out:
	for len(commitsBatch) != 0 {

		for _, commit := range commitsBatch {

			// Check if we already have this commit
			fromDB, err := repo.GetOneByParam("Commits", models.SnapshotParam{Name: "id", Value: commit.(map[string]interface{})["node_id"].(string)}, task.Groups)

			// If we already have this commit, stop the entire process
			// because the rest should be already in the database
			if err != nil || fromDB.ID != 0 || fromDB.Metric != "" {
				break out
			}

			// Get the commit detailed from GitHub API
			commitDetailed, err := getCommitDetailed(endpoint+"commits/"+commit.(map[string]interface{})["sha"].(string), apiKeys)
			if err != nil {
				break out
			}

			// Save files separately as a CommitFiles task snapshot
			// So that we don't send too much unnecessary data to the client
			commitFiles := commitDetailed.(map[string]interface{})["files"].([]interface{})
			for i, file := range commitFiles {
				file.(map[string]interface{})["commit_sha"] = commit.(map[string]interface{})["sha"]

				// Save the file data before deleting it in the commitDetailed
				fileJson, err := json.Marshal(file)
				if err != nil {
					continue
				}
				files = append(files, fileJson)

				// Remove unnecessary fields (relative to the commit snapshot)
				delete(commitDetailed.(map[string]interface{})["files"].([]interface{})[i].(map[string]interface{}), "blob_url")
				delete(commitDetailed.(map[string]interface{})["files"].([]interface{})[i].(map[string]interface{}), "commit_sha")
				delete(commitDetailed.(map[string]interface{})["files"].([]interface{})[i].(map[string]interface{}), "raw_url")
				delete(commitDetailed.(map[string]interface{})["files"].([]interface{})[i].(map[string]interface{}), "contents_url")
				delete(commitDetailed.(map[string]interface{})["files"].([]interface{})[i].(map[string]interface{}), "patch")

			}

			commits = append(commits, commitDetailed)
		}

		saveCommitsDetailed(commits, files, task, repo)

		commits = nil
		files = nil

		page += 1
		commitsBatch = getCommitBatch(endpoint, page, apiKeys)
	}

	saveCommitsDetailed(commits, files, task, repo)
}
