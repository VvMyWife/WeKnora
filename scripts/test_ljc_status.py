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

    def test_missing_tracking_is_reported(self):
        self.git("rm", "--cached", "frontend-custom/custom.txt")
        self.assertIn("frontend-custom/custom.txt", module.inspect(self.root)["changed_frontend_paths"])

    def test_approved_official_tree_only(self):
        (self.root / "frontend/original.txt").write_text("approved")
        self.git("add", "frontend/original.txt")
        self.git("commit", "-qm", "approved upstream tree")
        policy = self.root / "docs/upstream-integration"
        policy.mkdir(parents=True)
        (policy / "policy.json").write_text(json.dumps({"official_frontend_candidates": [self.git("rev-parse", "HEAD")]}))
        self.assertTrue(module.inspect(self.root)["frontend_frozen"])
        (self.root / "frontend/original.txt").write_text("unapproved")
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])

    def test_approved_override_is_pinned_by_hash(self):
        policy = self.root / "docs/upstream-integration"
        policy.mkdir(parents=True)
        ref = self.git("rev-parse", "HEAD")
        (self.root / "frontend/original.txt").write_text("mapped PDF support")
        (policy / "policy.json").write_text(json.dumps({
            "official_frontend_candidates": [ref],
            "official_frontend_overrides": {
                "frontend/original.txt": hashlib.sha256(b"mapped PDF support").hexdigest()
            }
        }))
        self.assertTrue(module.inspect(self.root)["frontend_frozen"])
        (self.root / "frontend/original.txt").write_text("unapproved replacement")
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])

    def test_custom_candidate_requires_approval_and_exact_content(self):
        policy_dir = self.root / "docs/upstream-integration"
        policy_dir.mkdir(parents=True)
        custom = self.root / "frontend-custom/custom.txt"
        custom.write_text("approved custom adaptation")
        manifest = {"files": [{"path": "frontend-custom/custom.txt",
                               "sha256": hashlib.sha256(custom.read_bytes()).hexdigest()}]}
        (policy_dir / "candidate.json").write_text(json.dumps(manifest))
        policy = {"custom_frontend_manifest": "docs/upstream-integration/candidate.json"}
        (policy_dir / "policy.json").write_text(json.dumps(policy))
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])
        policy["custom_frontend_approval"] = {"scope": ["all six modules"]}
        (policy_dir / "policy.json").write_text(json.dumps(policy))
        self.assertTrue(module.inspect(self.root)["frontend_frozen"])
        custom.write_text("unexpected later edit")
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])
        custom.unlink()
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])

    def test_custom_candidate_rejects_unpinned_new_file(self):
        policy_dir = self.root / "docs/upstream-integration"
        policy_dir.mkdir(parents=True)
        freeze = json.loads((self.root / "docs/fork-maintenance/frontend-freeze.json").read_text())
        (policy_dir / "candidate.json").write_text(json.dumps(freeze))
        (policy_dir / "policy.json").write_text(json.dumps({
            "custom_frontend_manifest": "docs/upstream-integration/candidate.json",
            "custom_frontend_approval": {"scope": ["all six modules"]},
        }))
        (self.root / "frontend-custom/new.txt").write_text("not pinned")
        self.git("add", "frontend-custom/new.txt")
        self.assertFalse(module.inspect(self.root)["frontend_frozen"])

    def test_interrupted_cherry_pick_is_reported(self):
        (self.root / ".git/CHERRY_PICK_HEAD").write_text(self.git("rev-parse", "HEAD"))
        result = module.inspect(self.root)
        self.assertEqual(["CHERRY_PICK_HEAD"], result["pending_git_operations"])
        self.assertIn("Never replay", result["next_action"])


if __name__ == "__main__":
    unittest.main()
