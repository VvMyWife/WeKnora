import tempfile
from pathlib import Path
import unittest

from export_processing_logs import atomic_write, format_report, one_line
from install_processing_log_cron import MARKER, updated_crontab


class ProcessingLogExportTests(unittest.TestCase):
    def test_one_line_removes_control_spacing(self):
        self.assertEqual(one_line("first\nsecond\tthird"), "first second third")

    def test_report_lists_counts_document_and_stage_error(self):
        docs = [
            {"id": "a", "folder_path": "folder", "file_name": "a.md",
             "parse_status": "finalizing", "summary_status": "failed",
             "error_message": "model timeout"},
            {"id": "b", "folder_path": "", "file_name": "b.md",
             "parse_status": "completed", "summary_status": "completed"},
        ]
        failures = [{"updated_at": "2026-10-14", "folder_path": "folder",
                     "file_name": "a.md", "attempt": 2, "name": "postprocess.wiki",
                     "error_code": "EXTRACT_FAILED", "error_message": "bad JSON",
                     "error_detail": "bad JSON at offset 3", "total_failures": 1}]
        stages = [{"stage": "postprocess.graph", "status": "failed", "count": 1}]
        report = format_report("kb", "政策库", docs, failures, stages, "now")
        for expected in ("文档总数：2", "当前失败文档：1", "folder/a.md",
                         "EXTRACT_FAILED", "bad JSON at offset 3", "postprocess.graph：failed 1"):
            self.assertIn(expected, report)

    def test_atomic_write_replaces_file(self):
        with tempfile.TemporaryDirectory() as temporary:
            target = Path(temporary) / "status.txt"
            atomic_write(target, "old")
            atomic_write(target, "new")
            self.assertEqual(target.read_text(encoding="utf-8"), "new")

    def test_cron_install_preserves_unrelated_jobs_and_is_idempotent(self):
        existing = "0 1 * * * /usr/bin/other-job\n"
        installed = updated_crontab(existing, remove=False)
        self.assertEqual(updated_crontab(installed, remove=False), installed)
        self.assertIn(MARKER, installed)
        self.assertEqual(updated_crontab(installed, remove=True), existing)


if __name__ == "__main__":
    unittest.main()
