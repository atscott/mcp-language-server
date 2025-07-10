package references_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isaacphi/mcp-language-server/integrationtests/tests/common"
	"github.com/isaacphi/mcp-language-server/integrationtests/tests/rust/internal"
	"github.com/isaacphi/mcp-language-server/internal/tools"
)

// TestFindReferences tests the FindReferences tool with Rust symbols
// that have references across different files
func TestFindReferences(t *testing.T) {
	// Helper function to open all files and wait for indexing
	openAllFilesAndWait := func(suite *common.TestSuite, ctx context.Context) {
		// Open all files to ensure rust-analyzer indexes everything
		filesToOpen := []string{
			"src/main.rs",
			"src/types.rs",
			"src/helper.rs",
			"src/consumer.rs",
			"src/another_consumer.rs",
			"src/clean.rs",
		}

		for _, file := range filesToOpen {
			filePath := filepath.Join(suite.WorkspaceDir, file)
			err := suite.Client.OpenFile(ctx, filePath)
			if err != nil {
				// Don't fail the test, some files might not exist in certain tests
				t.Logf("Note: Failed to open %s: %v", file, err)
			}
		}
	}

	suite := internal.GetTestSuite(t)

	ctx, cancel := context.WithTimeout(suite.Context, 10*time.Second)
	defer cancel()

	// Open all files and wait for rust-analyzer to index them
	openAllFilesAndWait(suite, ctx)

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
			filePath:      "src/helper.rs",
			line:          4,
			column:        8,
			expectedText:  "helper_function",
			expectedFiles: 2,
			snapshotName:  "helper-function",
			symbolForLog:  "helper_function",
		},
		{
			name:          "Function with reference in same file",
			filePath:      "src/main.rs",
			line:          9,
			column:        4,
			expectedText:  "main()",
			expectedFiles: 1, // main.rs
			snapshotName:  "foobar-function",
			symbolForLog:  "foo_bar",
		},
		{
			name:          "Struct with references across files",
			filePath:      "src/types.rs",
			line:          54,
			column:        8,
			expectedText:  "consumer_function",
			expectedFiles: 2, // consumer.rs and another_consumer.rs
			snapshotName:  "shared-struct",
			symbolForLog:  "SharedStruct",
		},
		{
			name:          "Method with references across files",
			filePath:      "src/types.rs",
			line:          64,
			column:        8,
			expectedText:  "method",
			expectedFiles: 1,
			snapshotName:  "struct-method",
			symbolForLog:  "method",
		},
		{
			name:          "Interface with references across files",
			filePath:      "src/types.rs",
			line:          70,
			column:        8,
			expectedText:  "iface",
			expectedFiles: 2, // consumer.rs and another_consumer.rs
			snapshotName:  "shared-interface",
			symbolForLog:  "SharedInterface",
		},
		{
			name:          "Interface method with references",
			filePath:      "src/types.rs",
			line:          71,
			column:        5,
			expectedText:  "get_name",
			expectedFiles: 2,
			snapshotName:  "interface-method",
			symbolForLog:  "get_name",
		},
		{
			name:          "Constant with references across files",
			filePath:      "src/types.rs",
			line:          81,
			column:        8,
			expectedText:  "SHARED_CONSTANT",
			expectedFiles: 2,
			snapshotName:  "shared-constant",
			symbolForLog:  "SHARED_CONSTANT",
		},
		{
			name:          "Type with references across files",
			filePath:      "src/types.rs",
			line:          79,
			column:        8,
			expectedText:  "SharedType",
			expectedFiles: 2,
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
			common.SnapshotTest(t, "rust", "references", tc.snapshotName, result)
		})
	}
}

// countFilesInResult counts the number of unique files mentioned in the result
func countFilesInResult(result string) int {
	fileMap := make(map[string]bool)

	// Any line containing "workspace" and ".rs" is a file path
	for line := range strings.SplitSeq(result, "\n") {
		if strings.Contains(line, "workspace") && strings.Contains(line, ".rs") {
			fileMap[line] = true
		}
	}

	return len(fileMap)
}
