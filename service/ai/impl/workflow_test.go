package aiimpl

import (
	"context"
	"strings"
	"testing"

	ai "github.com/evo-lee/evo-sonic/service/ai"
)

func svc() ai.WorkflowService { return NewWorkflowService() }

func TestSEOCheck_PerfectPost_Score100(t *testing.T) {
	report, err := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          1,
		Title:           "How to Build a Blog with Go and Hertz",
		Summary:         strings.Repeat("a", 80),
		OriginalContent: strings.Repeat("word ", 100),
		MetaDescription: "A concise meta description",
		TagCount:        4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Score != 100 {
		t.Errorf("want 100 got %d; suggestions: %+v", report.Score, report.Suggestions)
	}
	if len(report.Suggestions) != 0 {
		t.Errorf("want no suggestions, got %+v", report.Suggestions)
	}
}

func TestSEOCheck_EmptyPost_LowScore(t *testing.T) {
	report, err := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{PostID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if report.Score >= 50 {
		t.Errorf("empty post should score < 50, got %d", report.Score)
	}
}

func TestSEOCheck_EmptyTitle_ErrorSeverity(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          3,
		Summary:         strings.Repeat("x", 60),
		OriginalContent: strings.Repeat("x", 400),
		TagCount:        3,
	})
	hasTitleError := false
	for _, s := range report.Suggestions {
		if s.Field == "title" && s.Severity == "error" {
			hasTitleError = true
		}
	}
	if !hasTitleError {
		t.Error("expected title error suggestion for empty title")
	}
}

func TestSEOCheck_TitleTooLong_Warning(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          4,
		Title:           strings.Repeat("x", 75),
		Summary:         strings.Repeat("x", 60),
		OriginalContent: strings.Repeat("x", 400),
		TagCount:        3,
	})
	for _, s := range report.Suggestions {
		if s.Field == "title" && s.Severity == "warning" {
			return
		}
	}
	t.Error("expected title too-long warning")
}

func TestSEOCheck_NoTags_Warning(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          5,
		Title:           "Good Title Here",
		Summary:         strings.Repeat("x", 60),
		OriginalContent: strings.Repeat("x", 400),
		TagCount:        0,
	})
	for _, s := range report.Suggestions {
		if s.Field == "tags" && s.Severity == "warning" {
			return
		}
	}
	t.Error("expected tags warning for zero tags")
}

func TestSEOCheck_NoMetaDescription_Info(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          6,
		Title:           "Good Title Here",
		Summary:         strings.Repeat("x", 60),
		OriginalContent: strings.Repeat("x", 400),
		TagCount:        4,
		MetaDescription: "",
	})
	for _, s := range report.Suggestions {
		if s.Field == "meta_description" {
			return
		}
	}
	t.Error("expected meta_description suggestion")
}

func TestSEOCheck_ShortContent_Warning(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{
		PostID:          7,
		Title:           "Good Title",
		Summary:         strings.Repeat("x", 60),
		OriginalContent: "Short.",
		TagCount:        3,
	})
	for _, s := range report.Suggestions {
		if s.Field == "content" && s.Severity == "warning" {
			return
		}
	}
	t.Error("expected content too-short warning")
}

func TestSEOCheck_ScoreNeverNegative(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{PostID: 8})
	if report.Score < 0 {
		t.Errorf("score must be >= 0, got %d", report.Score)
	}
}

func TestSEOCheck_PostIDPreserved(t *testing.T) {
	report, _ := svc().CheckSEO(context.Background(), ai.SEOCheckRequest{PostID: 42})
	if report.PostID != 42 {
		t.Errorf("post_id not preserved: want 42 got %d", report.PostID)
	}
}
