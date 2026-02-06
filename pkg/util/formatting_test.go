package util

import (
	"testing"
)

func TestGetResourceColumns(t *testing.T) {
	tests := []struct {
		resourceType string
		expectedLen  int
		expectedCols []string
	}{
		{
			resourceType: "pods",
			expectedLen:  4,
			expectedCols: []string{"READY", "STATUS", "RESTARTS", "AGE"},
		},
		{
			resourceType: "services",
			expectedLen:  5,
			expectedCols: []string{"TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"},
		},
		{
			resourceType: "deployments",
			expectedLen:  4,
			expectedCols: []string{"READY", "UP-TO-DATE", "AVAILABLE", "AGE"},
		},
		{
			resourceType: "unknown-resource",
			expectedLen:  1,
			expectedCols: []string{"AGE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			cols := GetResourceColumns(tt.resourceType)

			if len(cols) != tt.expectedLen {
				t.Errorf("GetResourceColumns(%s) len = %d, want %d", tt.resourceType, len(cols), tt.expectedLen)
			}

			for i, expectedCol := range tt.expectedCols {
				if i >= len(cols) {
					t.Errorf("GetResourceColumns(%s) missing column at index %d", tt.resourceType, i)
					continue
				}
				if cols[i].Name != expectedCol {
					t.Errorf("GetResourceColumns(%s)[%d].Name = %s, want %s", tt.resourceType, i, cols[i].Name, expectedCol)
				}
			}
		})
	}
}
