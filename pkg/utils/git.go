package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitInfo 保存git信息
type GitInfo struct {
	CommitHash string
	CommitMsg  string
	Branch     string
}

// GetGitInfo 获取当前git仓库的信息
func GetGitInfo() (*GitInfo, error) {
	// 获取最后一次提交的哈希值
	hash, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git commit hash: %v", err)
	}

	// 获取最后一次提交的消息
	msg, err := exec.Command("git", "log", "-1", "--pretty=%B").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git commit message: %v", err)
	}

	// 获取当前分支
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git branch: %v", err)
	}

	return &GitInfo{
		CommitHash: strings.TrimSpace(string(hash)),
		CommitMsg:  strings.TrimSpace(string(msg)),
		Branch:     strings.TrimSpace(string(branch)),
	}, nil
}
