package references_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/python/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestFindReferences tests the FindReferences tool with Python symbols
// that have references across different files
func TestFindReferences(t *testing.T) {
	suite := internal.GetTestSuite(t)

	ctx, cancel := context.WithTimeout(suite.Context, 10*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		filePath      string
		line          int
		column        int
		expectedText  string
		expectedFiles int // Number of files where references should be found
		snapshotName  string
		symbolForLog  string
	}{
		{
			name:          "Function with references across files",
			filePath:      "helper.py",
			line:          80,
			column:        5,
			expectedText:  "helper_function",
			expectedFiles: 2, // consumer.py and another_consumer.py
			snapshotName:  "helper-function",
			symbolForLog:  "helper_function",
		},
		{
			name:          "Class with references across files",
			filePath:      "helper.py",
			line:          28,
			column:        7,
			expectedText:  "SharedClass",
			expectedFiles: 2, // consumer.py and another_consumer.py
			snapshotName:  "shared-class",
			symbolForLog:  "SharedClass",
		},
		{
			name:          "Method with references across files",
			filePath:      "helper.py",
			line:          43,
			column:        9,
			expectedText:  "get_name",
			expectedFiles: 2, // consumer.py and another_consumer.py
			snapshotName:  "class-method",
			symbolForLog:  "get_name",
		},
		{
			name:          "Interface with references across files",
			filePath:      "helper.py",
			line:          63,
			column:        7,
			expectedText:  "SharedInterface",
			expectedFiles: 1, // consumer.py
			snapshotName:  "shared-interface",
			symbolForLog:  "SharedInterface",
		},
		{
			name:          "Interface method with references",
			filePath:      "helper.py",
			line:          66,
			column:        9,
			expectedText:  "process",
			expectedFiles: 1, // consumer.py
			snapshotName:  "interface-method",
			symbolForLog:  "process",
		},
		{
			name:          "Constant with references across files",
			filePath:      "helper.py",
			line:          8,
			column:        1,
			expectedText:  "SHARED_CONSTANT",
			expectedFiles: 2, // consumer.py and another_consumer.py
			snapshotName:  "shared-constant",
			symbolForLog:  "SHARED_CONSTANT",
		},
		{
			name:          "Enum-like class with references across files",
			filePath:      "helper.py",
			line:          12,
			column:        7,
			expectedText:  "Color",
			expectedFiles: 2, // consumer.py and another_consumer.py
			snapshotName:  "color-enum",
			symbolForLog:  "Color",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(suite.WorkspaceDir, tc.filePath)
			// Call the FindReferences tool
			result, err := tools.FindReferences(ctx, suite.Client, filePath, tc.line, tc.column)
			if err != nil {
				t.Fatalf("Failed to find references for %s: %v", tc.symbolForLog, err)
			}

			// Check that the result contains relevant information
			if !strings.Contains(result, tc.expectedText) {
				t.Errorf("References do not contain expected text: %s", tc.expectedText)
			}

			// Count how many different files are mentioned in the result
			fileCount := countFilesInResult(result)
			if fileCount < tc.expectedFiles {
				t.Errorf("Expected references in at least %d files, but found in %d files",
					tc.expectedFiles, fileCount)
			}

			// Use snapshot testing to verify exact output
			common.SnapshotTest(t, "python", "references", tc.snapshotName, result)
		})
	}
}

// countFilesInResult counts the number of unique files mentioned in the result
func countFilesInResult(result string) int {
	fileMap := make(map[string]bool)

	// Any line containing "workspace" and ".py" is a file path
	for line := range strings.SplitSeq(result, "\n") {
		if strings.Contains(line, "workspace") && strings.Contains(line, ".py") {
			if !strings.Contains(line, "References in File") {
				fileMap[line] = true
			}
		}
	}

	return len(fileMap)
}
