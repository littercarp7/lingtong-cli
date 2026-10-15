// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

func newCmdTableDataBatchCreate(f *cmdutil.Factory) *cobra.Command {
	var basicDataID, schemaID, version int
	var records, recordsFile string
	cmd := &cobra.Command{
		Use:   "batch-create",
		Short: "Batch create or upsert table records",
		Long: `Send a JSON array of records to POST /basicdata/record/batchSave.

Each record needs a non-empty data object keyed by field ID. The table and
schema IDs are filled from flags; --version fills only records without a version.
Use --records-file for a JSON array converted from an Excel worksheet.

The server checks primary keys: duplicate keys within a batch fail, and keys
already in the table are updated with all fields (omitted fields may be cleared).
The request is sent as a single batch with no automatic splitting, retries,
progress, or checkpointing. For large imports, submit sequential batches and
save a checkpoint only after each successful command.

EXAMPLES:
    lingtong-cli table data batch-create --basic-data-id 1633 --schema-id 2546 \
      --records '[{"data":{"1":"产品A","2":10}},{"data":{"1":"产品B","2":20}}]'

    lingtong-cli table data batch-create --basic-data-id 1633 --schema-id 2546 \
      --version 5 --records-file ./records.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if basicDataID <= 0 || schemaID <= 0 {
				return fmt.Errorf("--basic-data-id and --schema-id must be positive")
			}
			if (records == "") == (recordsFile == "") {
				return fmt.Errorf("provide only one of --records or --records-file")
			}
			if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
				return err
			}

			source := []byte(records)
			if recordsFile != "" {
				var err error
				source, err = os.ReadFile(recordsFile)
				if err != nil {
					return fmt.Errorf("read --records-file: %w", err)
				}
			}
			var recordList []map[string]interface{}
			if err := json.Unmarshal(source, &recordList); err != nil {
				return fmt.Errorf("invalid --records JSON array: %w", err)
			}
			if len(recordList) == 0 {
				return fmt.Errorf("--records must contain at least one record")
			}
			for i, rec := range recordList {
				data, ok := rec["data"].(map[string]interface{})
				if !ok || len(data) == 0 {
					return fmt.Errorf("record %d: data must be a non-empty JSON object", i+1)
				}
				if id, exists := rec["basicDataId"]; exists && id != float64(basicDataID) {
					return fmt.Errorf("record %d: basicDataId does not match --basic-data-id", i+1)
				}
				if id, exists := rec["schemaId"]; exists && id != float64(schemaID) {
					return fmt.Errorf("record %d: schemaId does not match --schema-id", i+1)
				}
				rec["basicDataId"] = basicDataID
				rec["schemaId"] = schemaID
				if version > 0 {
					if _, exists := rec["version"]; !exists {
						rec["version"] = version
					}
				}
			}

			c := client.NewClient(f.Config.Host, f.Config.Token)
			resp, err := c.Post("/basicdata/record/batchSave", recordList)
			if err != nil {
				return fmt.Errorf("failed to batch create records: %w", err)
			}
			result, err := parseTableMutationResponse(resp, "batch create", len(recordList))
			if err != nil {
				return err
			}
			return f.NewWriter(output.Format(cmd.Flag("format").Value.String())).Write(result)
		},
	}
	cmd.Flags().IntVar(&basicDataID, "basic-data-id", 0, "Table ID (required)")
	cmd.Flags().IntVar(&schemaID, "schema-id", 0, "Schema ID (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Schema version (optional, applied to records without version)")
	cmd.Flags().StringVar(&records, "records", "", "JSON array of records")
	cmd.Flags().StringVar(&recordsFile, "records-file", "", "Path to a JSON array of records")
	_ = cmd.MarkFlagRequired("basic-data-id")
	_ = cmd.MarkFlagRequired("schema-id")
	return cmd
}
