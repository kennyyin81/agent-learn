package agents

import "workflow-multiagent-demo/internal/shared"

/**
 * 计划结构
 * Goal: 目标
 * Tasks: 任务列表
 * Constraints: 约束条件
 * ExpectedFiles: 期望生成的文件列表
**/
type Plan struct {
	Goal          string   `json:"goal"`
	Tasks         []string `json:"tasks"`
	Constraints   []string `json:"constraints"`
	ExpectedFiles []string `json:"expected_files"`
}

/**
 * 代码文件结构
 * Path: 文件路径
 * Content: 文件内容
**/
type CodeFile = shared.CodeFile

/**
 * 代码结果结构
 * Files: 生成的文件列表
 * Notes: 备注信息
**/
type CodeResult struct {
	Files []CodeFile `json:"files"`
	Notes string     `json:"notes"`
}

/**
 * 代码审查结果结构
 * Passed: 审查是否通过
 * Issues: 发现的问题列表
 * Suggestions: 改进建议列表
 * RepairInstruction: 修复指导
**/
type ReviewResult struct {
	Passed            bool     `json:"passed"`
	Issues            []string `json:"issues"`
	Suggestions       []string `json:"suggestions"`
	RepairInstruction string   `json:"repair_instruction"`
}

/**
 * 最终报告结构
 * Summary: 摘要
 * Files: 涉及的文件列表
 * TestResult: 测试结果
 * Review: 审查意见
**/
type FinalReport struct {
	Summary    string   `json:"summary"`
	Files      []string `json:"files"`
	TestResult string   `json:"test_result"`
	Review     string   `json:"review"`
	NextSteps  []string `json:"next_steps"`
}
