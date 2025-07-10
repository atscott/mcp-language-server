package references_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/go/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestFindReferences tests the FindReferences tool with Go symbols
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
			filePath:      "helper.go",
			line:          4,
			column:        6,
			expectedText:  "ConsumerFunction",
			expectedFiles: 2, // consumer.go and another_consumer.go
			snapshotName:  "helper-function",
			symbolForLog:  "HelperFunction",
		},
		{
			name:          "Function with reference in same file",
			filePath:      "main.go",
			line:          6,
			column:        6,
			expectedText:  "main()",
			expectedFiles: 1, // main.go
			snapshotName:  "foobar-function",
			symbolForLog:  "FooBar",
		},
		{
			name:          "Struct with references across files",
			filePath:      "types.go",
			line:          6,
			column:        6,
			expectedText:  "ConsumerFunction",
			expectedFiles: 2, // consumer.go and another_consumer.go
			snapshotName:  "shared-struct",
			symbolForLog:  "SharedStruct",
		},
		{
			name:          "Method with references across files",
			filePath:      "types.go",
			line:          15,
			column:        24,
			expectedText:  "s.Method()",
			expectedFiles: 1, // consumer.go
			snapshotName:  "struct-method",
			symbolForLog:  "SharedStruct.Method",
		},
		{
			name:          "Interface with references across files",
			filePath:      "types.go",
			line:          20,
			column:        6,
			expectedText:  "var iface SharedInterface",
			expectedFiles: 2, // consumer.go and another_consumer.go
			snapshotName:  "shared-interface",
			symbolForLog:  "SharedInterface",
		},
		{
			name:          "Interface method with references",
			filePath:      "types.go",
			line:          22,
			column:        2,
			expectedText:  "iface.GetName()",
			expectedFiles: 1, // consumer.go
			snapshotName:  "interface-method",
			symbolForLog:  "SharedInterface.GetName",
		},
		{
			name:          "Constant with references across files",
			filePath:      "types.go",
			line:          26,
			column:        7,
			expectedText:  "SharedConstant",
			expectedFiles: 2, // consumer.go and another_consumer.go
			snapshotName:  "shared-constant",
			symbolForLog:  "SharedConstant",
		},
		{
			name:          "Type with references across files",
			filePath:      "types.go",
			line:          29,
			column:        6,
			expectedText:  "SharedType",
			expectedFiles: 2, // consumer.go and another_consumer.go
			snapshotName:  "shared-type",
			symbolForLog:  "SharedType",
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
			common.SnapshotTest(t, "go", "references", tc.snapshotName, result)
		})
	}
}

// countFilesInResult counts the number of unique files mentioned in the result
func countFilesInResult(result string) int {
	fileMap := make(map[string]bool)

	// Any line containing "workspace" and ".go" is a file path
	for line := range strings.SplitSeq(result, "\n") {
		if strings.Contains(line, "workspace") && strings.Contains(line, ".go") {
			if !strings.Contains(line, "References in File") {
				fileMap[line] = true
			}
		}
	}

	return len(fileMap)
}
