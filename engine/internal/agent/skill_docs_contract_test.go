package agent

// skill 文档 ↔ 引擎 schema 契约测试：
//  1. liki-skills 文档（SKILL.md/app/domains）反引号引用的字段名必须存在于某方法
//     Params/Result 的 JSON Schema 属性集合。
//  2. skill 侧 check_docs.py 的方法白名单 == 引擎注册方法集。
//
// 依赖本 monorepo 的 skills/ 目录（相对路径基于包目录 engine/internal/agent：../../../skills）；目录缺失时自动跳过。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// 校验范围：liki-bazi（规则引擎）+ 子流程三 skill（domains 表引用引擎字段——jixiong/yiji 等）
// 相对路径基于包目录 engine/internal/agent（go test cwd）：../../../skills/... 指向本 monorepo 根 skills/
var skillDocsRels = []string{
	"../../../skills/liki-bazi",
	"../../../skills/liki-divination",
	"../../../skills/liki-fengshui",
	"../../../skills/liki-naming",
}

var (
	reFieldToken = regexp.MustCompile(`^[a-zA-Z_][\w.\[\]/]*$`)
	reEnvVar     = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	reExtension  = regexp.MustCompile(`\.(md|py|json|csv|sh|yaml|yml|toml|txt)$`)
)

// fieldRefs 收集方法 schema 属性路径：type=object→properties 递归；type=array→items 递归（路径记 []）。
type fieldRefs struct {
	paths  []string        // 完整属性路径（相对 schema 根，如 "gong_wei[].xing_yao"）
	leaves map[string]bool // 全部叶子属性名
}

func collectFieldRefs(schema json.RawMessage, root string, out *fieldRefs) {
	var doc any
	if len(schema) == 0 || json.Unmarshal(schema, &doc) != nil {
		return
	}
	var walk func(node any, path string)
	walk = func(node any, path string) {
		obj, ok := node.(map[string]any)
		if !ok {
			return
		}
		// array 节点：遍历 items（顶层 data 为数组的 Result——如 huangli.days 返回数组）
		if obj["type"] == "array" {
			if items, ok := obj["items"].(map[string]any); ok {
				walk(items, path)
			}
			return
		}
		if props, ok := obj["properties"].(map[string]any); ok {
			for name, sub := range props {
				p := path + name
				out.leaves[name] = true
				if subMap, ok := sub.(map[string]any); ok {
					switch subMap["type"] {
					case "object":
						walk(subMap, p+".")
					case "array":
						if items, ok := subMap["items"].(map[string]any); ok {
							out.paths = append(out.paths, p+"[]")
							walk(items, p+"[].")
						} else {
							out.paths = append(out.paths, p+"[]")
						}
					default:
						out.paths = append(out.paths, p)
					}
				} else {
					out.paths = append(out.paths, p)
				}
			}
		}
	}
	// 跳过 envelope（_product/data）——字段引用指 data 内层
	if obj, ok := doc.(map[string]any); ok {
		if props, ok := obj["properties"].(map[string]any); ok {
			if data, ok := props["data"]; ok {
				walk(data, root)
				return
			}
		}
	}
	walk(doc, root)
}

func registryFieldRefs(reg *RPCRegistry) *fieldRefs {
	out := &fieldRefs{leaves: map[string]bool{}}
	for name, m := range reg.methods {
		collectFieldRefs(m.Params, "params.", out)
		collectFieldRefs(m.Result, name+".", out)
	}
	return out
}

// 字段 token 规范化：去 [] 空括号、按 / 拆分枚举（start_date/end_date → 逐段校验）。
func normalizeFieldToken(tok string) []string {
	tok = strings.ReplaceAll(tok, "[]", "")
	tok = regexp.MustCompile(`\[\d+\]`).ReplaceAllString(tok, "") // 数组索引访问（wang_shuai[2] → wang_shuai）
	if strings.Contains(tok, "/") {
		var out []string
		for _, seg := range strings.Split(tok, "/") {
			if seg != "" {
				out = append(out, seg)
			}
		}
		return out
	}
	return []string{tok}
}

// pathResolvable：规范化 token 能否在字段集合中解析——完整路径匹配 或 逐段前缀匹配。
func pathResolvable(tok string, refs *fieldRefs) bool {
	if refs.leaves[tok] {
		return true
	}
	// 路径逐段解析：da_yun.steps → 存在以 "da_yun.steps" 结尾的路径
	segs := strings.Split(tok, ".")
	for i := 1; i < len(segs); i++ {
		prefix := strings.Join(segs[:i+1], ".")
		for _, p := range refs.paths {
			norm := strings.ReplaceAll(p, "[]", "")
			if norm == prefix || strings.Contains(norm, "."+prefix) {
				return true
			}
		}
	}
	return false
}

func loadSkillDocs() ([]string, error) {
	var files []string
	for _, root := range skillDocsRels {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func TestSkillDocsFieldRefs(t *testing.T) {
	files, err := loadSkillDocs()
	if err != nil || len(files) == 0 {
		t.Skip("skills/ 目录不存在，跳过 skill 文档契约测试")
	}
	reg := NewRPCRegistry()
	refs := registryFieldRefs(reg)
	// 放行集合：引擎方法名 + skill 工具名（full_paipan 等非引擎 RPC 字段）+ OpenRPC 文档字段 + skill 根文件
	allow := map[string]bool{}
	for name := range reg.methods {
		allow[name] = true
	}
	for _, a := range []string{"rpc.discover", "full_paipan", "city_coords",
		"query", "yearly_range", "calibrate", "bond", "skill-tools.json",
		"VERSION", "content.sha256", "liki-memory.json", "RPCError", "ValueError",
		"error",
		"methods", "parameters", "required", "params.properties", "result.methods"} {
		allow[a] = true
	}

	var unresolved []string
	lineToken := regexp.MustCompile("`([^`]+)`")
	for _, f := range files {
		raw, _ := os.ReadFile(f)
		// 逐行扫描：字段引用均为单行内成对反引号；含奇数反引号的行（```json 代码块边界）跳过，
		// 避免三反引号代码块与单反引号配对错位（markdown 嵌套导致 findall 跨行吞 token）
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.Count(line, "`")%2 != 0 {
				continue
			}
			for _, bt := range lineToken.FindAllStringSubmatch(line, -1) {
				tok := strings.TrimSpace(bt[1])
				if strings.ContainsAny(tok, "{}") || reEnvVar.MatchString(tok) || !reFieldToken.MatchString(tok) { // reEnvVar: 全大写下划线=环境变量（LIKI_RPC_URL），非 schema 字段
					continue
				}
				if reExtension.MatchString(tok) || strings.HasPrefix(tok, "tools/") ||
					strings.HasPrefix(tok, "app/") || strings.HasPrefix(tok, "domains/") ||
					strings.HasPrefix(tok, "webapp/") {
					continue
				}
				// 含 '/' 但非已知路径前缀 → HTTP 头/媒体类型等值（如 Content-Type: application/json），非 schema 字段，跳过
				if strings.Contains(tok, "/") &&
					!strings.HasPrefix(tok, "tools/") && !strings.HasPrefix(tok, "app/") &&
					!strings.HasPrefix(tok, "domains/") && !strings.HasPrefix(tok, "webapp/") &&
					!strings.HasPrefix(tok, "skills/") {
					continue
				}
				for _, seg := range normalizeFieldToken(tok) {
					if allow[seg] || pathResolvable(seg, refs) {
						continue
					}
					rel, _ := filepath.Rel(filepath.Join("..", "..", ".."), f)
					unresolved = append(unresolved, rel+":"+line+": `"+tok+"`")
				}
			}
		}
	}
	sort.Strings(unresolved)
	if len(unresolved) > 0 {
		t.Errorf("skill 文档引用了引擎 schema 不存在的字段（%d 处）：\n  %s",
			len(unresolved), strings.Join(unresolved, "\n  "))
	}
}

// 双向同步：skill 侧 check_docs.py 方法白名单 == 引擎注册方法集（防两处漂移）。
func TestSkillDocsMethodListSync(t *testing.T) {
	path := filepath.Join("..", "..", "..", "tests", "check_docs.py")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip("check_docs.py 不在本仓库，跳过方法集同步测试")
	}
	whitelist := map[string]bool{}
	// 只提取 METHOD_WHITELIST = {...} 块内的字符串（避免 _SKIP_DOTTED 混入），允许无点单名
	seg := string(raw)
	if idx := strings.Index(seg, "METHOD_WHITELIST = {"); idx >= 0 {
		seg = seg[idx:]
		if end := strings.Index(seg, "}\n"); end >= 0 {
			seg = seg[:end]
		}
	}
	for _, m := range regexp.MustCompile(`"([a-z]+(?:\.[a-z_]+)?)"`).FindAllStringSubmatch(seg, -1) {
		whitelist[m[1]] = true
	}
	delete(whitelist, "rpc.discover") // OpenRPC 文档方法，引擎 registry 不注册

	var regNames []string
	for name := range NewRPCRegistry().methods {
		regNames = append(regNames, name)
	}
	var missingFromSkill, missingFromEngine []string
	for _, n := range regNames {
		if !whitelist[n] {
			missingFromSkill = append(missingFromSkill, n)
		}
	}
	for n := range whitelist {
		if !contains(regNames, n) {
			missingFromEngine = append(missingFromEngine, n)
		}
	}
	sort.Strings(missingFromSkill)
	sort.Strings(missingFromEngine)
	if len(missingFromSkill) > 0 || len(missingFromEngine) > 0 {
		t.Errorf("check_docs.py 方法白名单与引擎方法集不同步：\n  引擎有而白名单缺: %v\n  白名单有而引擎无: %v",
			missingFromSkill, missingFromEngine)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
