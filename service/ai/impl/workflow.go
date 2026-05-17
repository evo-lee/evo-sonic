package aiimpl

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/evo-lee/evo-sonic/injection"
	ai "github.com/evo-lee/evo-sonic/service/ai"
)

func init() {
	injection.Provide(NewWorkflowService)
}

type workflowServiceImpl struct{}

func NewWorkflowService() ai.WorkflowService {
	return &workflowServiceImpl{}
}

func (s *workflowServiceImpl) CheckSEO(_ context.Context, req ai.SEOCheckRequest) (*ai.SEOReport, error) {
	var suggestions []ai.SEOSuggestion
	deductions := 0

	// Title checks
	titleLen := utf8.RuneCountInString(req.Title)
	switch {
	case titleLen == 0:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "title", Severity: "error", Message: "标题不能为空"})
		deductions += 30
	case titleLen < 10:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "title", Severity: "warning", Message: "标题过短（建议 10-70 字）"})
		deductions += 10
	case titleLen > 70:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "title", Severity: "warning", Message: "标题过长（超过 70 字可能被搜索引擎截断）"})
		deductions += 5
	}

	// Summary checks
	summaryLen := utf8.RuneCountInString(strings.TrimSpace(req.Summary))
	switch {
	case summaryLen == 0:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "summary", Severity: "warning", Message: "缺少摘要（建议填写 50-160 字摘要）"})
		deductions += 15
	case summaryLen < 50:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "summary", Severity: "info", Message: "摘要较短（建议至少 50 字）"})
		deductions += 5
	case summaryLen > 160:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "summary", Severity: "info", Message: "摘要超过 160 字，搜索引擎可能截断"})
		deductions += 5
	}

	// Meta description
	if strings.TrimSpace(req.MetaDescription) == "" {
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "meta_description", Severity: "info", Message: "未设置 Meta 描述，建议填写"})
		deductions += 5
	}

	// Tag count checks
	switch {
	case req.TagCount == 0:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "tags", Severity: "warning", Message: "没有标签，建议添加 3-6 个相关标签"})
		deductions += 10
	case req.TagCount < 3:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "tags", Severity: "info", Message: "标签较少，建议添加 3-6 个标签"})
		deductions += 5
	case req.TagCount > 8:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "tags", Severity: "info", Message: "标签过多（超过 8 个），可能影响主题聚焦"})
		deductions += 3
	}

	// Content length (rough word/char count)
	contentLen := utf8.RuneCountInString(strings.TrimSpace(req.OriginalContent))
	switch {
	case contentLen == 0:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "content", Severity: "error", Message: "正文内容为空"})
		deductions += 30
	case contentLen < 300:
		suggestions = append(suggestions, ai.SEOSuggestion{Field: "content", Severity: "warning", Message: "正文较短（建议至少 300 字）"})
		deductions += 10
	}

	score := 100 - deductions
	if score < 0 {
		score = 0
	}

	return &ai.SEOReport{
		PostID:      req.PostID,
		Score:       score,
		Suggestions: suggestions,
	}, nil
}
