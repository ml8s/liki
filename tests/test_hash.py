"""content.sha256 指纹测试（外部评审 #1：版本相同但内容滞后，自检须能发现）。"""
import hashlib
import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'skills', 'liki-bazi', 'tools'))
import hash as hashmod


class TestContentHash(unittest.TestCase):
    def _tree(self, files):
        d = tempfile.mkdtemp()
        for rel, content in files.items():
            p = os.path.join(d, rel)
            os.makedirs(os.path.dirname(p), exist_ok=True)
            with open(p, 'w', encoding='utf-8') as f:
                f.write(content)
        return d

    def test_确定性_同树同hash(self):
        d = self._tree({"a.md": "x", "b/c.py": "y"})
        h1 = hashmod.tree_sha256(d)
        h2 = hashmod.tree_sha256(d)
        self.assertEqual(h1, h2)

    def test_内容变更_hash变(self):
        d1 = self._tree({"a.md": "x", "b/c.py": "y"})
        d2 = self._tree({"a.md": "x2", "b/c.py": "y"})
        self.assertNotEqual(hashmod.tree_sha256(d1), hashmod.tree_sha256(d2))

    def test_排除目录_ignored(self):
        d1 = self._tree({"a.md": "x", ".git/x": "g", "__pycache__/x.pyc": "p", "dist/x": "d"})
        d2 = self._tree({"a.md": "x"})
        self.assertEqual(hashmod.tree_sha256(d1), hashmod.tree_sha256(d2))

    def test_指纹可校验(self):
        d = self._tree({"SKILL.md": "s", "VERSION": "3.10.0"})
        fp = hashmod.tree_sha256(d)
        self.assertEqual(hashlib.sha256(fp.encode()).hexdigest()[:8], hashlib.sha256(fp.encode()).hexdigest()[:8])


if __name__ == "__main__":
    unittest.main()
