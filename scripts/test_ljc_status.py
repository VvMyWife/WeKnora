import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
import hashlib

spec = importlib.util.spec_from_file_location("ljc_status", Path(__file__).with_name("ljc_status.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class MaintenanceChecks(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.git("init", "-q")
        self.git("config", "user.name", "ljc")
        self.git("config", "user.email", "ljc@local")
        (self.root / "frontend").mkdir()
        (self.root / "frontend/original.txt").write_text("original")
        (self.root / "frontend-custom").mkdir()
        (self.root / "frontend-custom/custom.txt").write_text("custom")
        self.git("add", ".")
        self.git("commit", "-qm", "fixture")
        docs = self.root / "docs/fork-maintenance"
        docs.mkdir(parents=True)
        (docs / "baseline.json").write_text(json.dumps({"backend_head": self.git("rev-parse", "HEAD")}))
        (docs / "frontend-freeze.json").write_text(json.dumps({"files": [
            {"path": "frontend-custom/custom.txt", "sha256": hashlib.sha256(b"custom").hexdigest()}
        ]}))

    def git(self, *args):
        return subprocess.check_output(["git", "-C", str(self.root), *args], text=True).strip()

    def test_clean_frontends(self):
        self.assertTrue(module.inspect(self.root)["frontend_frozen"])

    def test_custom_edit_or_deletion(self):
        (self.root / "frontend-custom/custom.txt").write_text("changed")
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])
        (self.root / "frontend-custom/custom.txt").unlink()
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])

    def test_original_frontend_edit(self):
        (self.root / "frontend/original.txt").write_text("changed")
        self.assertIn("frontend/original.txt", module.inspect(self.root)["changed_frontend_paths"])

    def test_unexpected_tracked_frontend(self):
        (self.root / "frontend-custom/new.txt").write_text("new")
        self.git("add", "frontend-custom/new.txt")
        self.assertIn("frontend-custom/new.txt", module.inspect(self.root)["changed_frontend_paths"])

    def test_interrupted_cherry_pick_is_reported(self):
        (self.root / ".git/CHERRY_PICK_HEAD").write_text(self.git("rev-parse", "HEAD"))
        result = module.inspect(self.root)
        self.assertEqual(["CHERRY_PICK_HEAD"], result["pending_git_operations"])
        self.assertIn("Never replay", result["next_action"])


if __name__ == "__main__":
    unittest.main()
