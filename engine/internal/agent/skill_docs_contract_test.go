package agent

// skill 文档 ↔ 引擎 schema 契约测试：
//  1. liki-skills 文档（SKILL.md/app/domains）反引号引用的字段名必须存在于某方法
//     Params/Result 的 JSON Schema 属性集合。
//  2. skill 侧 check_docs.py 的方法白名单 == 引擎注册方法集。
//
// 依赖本 monorepo 的 skills/ 目录（相对路径基于包目录 engine/internal/agent：../../../skills）；目录缺失时自动跳过。

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// 校验范围：unified Liki skill（根入口 + 四个领域包）
// 相对路径基于包目录 engine/internal/agent（go test cwd）：../../../skills/... 指向本 monorepo 根 skills/
var skillDocsRels = []string{
	"../../../skills/liki",
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
		for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
			if branches, ok := obj[keyword].([]any); ok {
				for _, branch := range branches {
					walk(branch, path)
				}
			}
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
	// 放行集合：引擎方法名 + 当前 skill 工具 schema 词表 + OpenRPC 文档字段 + skill 根文件
	allow := map[string]bool{}
	for name := range reg.methods {
		allow[name] = true
		// MCP 工具名 = RPC 方法名点转下划线（bazi.chart → bazi_chart）
		allow[strings.ReplaceAll(name, ".", "_")] = true
	}
	toolVocabulary, err := loadSkillToolVocabulary()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []string{"rpc.discover", "skill-tools.json", "VERSION.txt", "content.sha256",
		"liki-memory.json", "RPCError", "ValueError", "error", "methods", "parameters",
		"required", "params.properties", "params.methods", "result.methods",
		"data", "result.data", "result.info", "result.info.version", "result_schema",
		"ok", "current_year", "current_year_source", "yong_shen_context", "years", "trace",
		"result.methods[].name", "bazhai.chart.result.data.ming_gua.gua.name",
		"isError", "mcpServers", "streamableHttp", "tools/call", "tools/list", "Mcp-Method",
		"result.methods.name", "xuankong.chart.result.data", "bazhai",
		"xuankong", "qiming", "snapshot", "data", "error", "unknown",
		"true", "false", "skills/liki/VERSION.txt", "create_birth_chart.data",
		"data.chart_ref", "error.code", "error.message", "code", "message",
		"safety_advisory", "meta.skill", "info.version", "pan_digest",
		"pan.ziwei_daxian"} {
		allow[a] = true
	}

	var unresolved []string
	lineToken := regexp.MustCompile("`([^`]+)`")
	for _, f := range files {
		docAllow := allow
		docSkill := skillNameForDoc(f)
		if docSkill == "liki" {
			docAllow = make(map[string]bool, len(allow))
			maps.Copy(docAllow, allow)
			for _, vocabulary := range toolVocabulary {
				maps.Copy(docAllow, vocabulary)
			}
		} else if vocabulary, exists := toolVocabulary[docSkill]; exists {
			docAllow = make(map[string]bool, len(allow)+len(vocabulary))
			maps.Copy(docAllow, allow)
			maps.Copy(docAllow, vocabulary)
		}
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
					strings.HasPrefix(tok, "webapp/") || strings.HasPrefix(tok, "skills/liki/") {
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
					if fieldAllowed(seg, docAllow, refs) {
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

func skillNameForDoc(path string) string {
	rel, err := filepath.Rel(filepath.Join("..", "..", "..", "skills"), path)
	if err != nil {
		return ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 0 {
		return ""
	}
	if parts[0] != "liki" {
		return parts[0]
	}
	if len(parts) == 1 || parts[len(parts)-1] == "SKILL.md" {
		return "liki"
	}
	return parts[1]
}

func loadSkillToolVocabulary() (map[string]map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "analysis", "liki_analysis", "*", "tools", "skill-tools.json"))
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]bool)
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var document struct {
			Tools []struct {
				Function struct {
					Name       string         `json:"name"`
					Parameters map[string]any `json:"parameters"`
				} `json:"function"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			return nil, err
		}
		parts := strings.Split(filepath.ToSlash(path), "/")
		skill := "liki"
		for index, part := range parts {
			if (part == "liki" || part == "liki_analysis") && index+1 < len(parts) {
				skill = parts[index+1]
				break
			}
		}
		vocabulary := result[skill]
		if vocabulary == nil {
			vocabulary = make(map[string]bool)
			result[skill] = vocabulary
		}
		for _, tool := range document.Tools {
			vocabulary[tool.Function.Name] = true
			collectToolVocabulary(tool.Function.Parameters, vocabulary)
		}

		// Response contracts are not sent to the LLM. Load them only for
		// documentation vocabulary so output fields stay externally documented.
		responsePath := filepath.Join(filepath.Dir(path), "response-contract.json")
		if responseRaw, err := os.ReadFile(responsePath); err == nil {
			var responseContract struct {
				Tools map[string]any `json:"tools"`
			}
			if err := json.Unmarshal(responseRaw, &responseContract); err != nil {
				return nil, err
			}
			for _, schema := range responseContract.Tools {
				collectToolVocabulary(schema, vocabulary)
			}
		}

		// The unified root owns the feedback contract. Keep its vocabulary under
		// the product skill so domain tool vocabularies stay scoped.
		feedbackPath := filepath.Join("..", "..", "..", "skills", "liki", "feedback.schema.json")
		if feedbackRaw, err := os.ReadFile(feedbackPath); err == nil {
			var feedbackSchema any
			if err := json.Unmarshal(feedbackRaw, &feedbackSchema); err != nil {
				return nil, err
			}
			if result["liki"] == nil {
				result["liki"] = make(map[string]bool)
			}
			collectToolVocabulary(feedbackSchema, result["liki"])
		}
	}

	// Skills without a Python tool layer still publish their root feedback schema.
	feedbackFiles, err := filepath.Glob(filepath.Join("..", "..", "..", "skills", "liki", "feedback.schema.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range feedbackFiles {
		skill := "liki"
		vocabulary := result[skill]
		if vocabulary == nil {
			vocabulary = make(map[string]bool)
			result[skill] = vocabulary
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var schema any
		if err := json.Unmarshal(raw, &schema); err != nil {
			return nil, err
		}
		collectToolVocabulary(schema, vocabulary)
	}
	return result, nil
}

func TestSkillToolVocabularyIsScoped(t *testing.T) {
	vocabulary, err := loadSkillToolVocabulary()
	if err != nil {
		t.Fatal(err)
	}
	natal := vocabulary["natal"]
	divination := vocabulary["divination"]
	if !natal["create_birth_chart"] || !natal["analyze_natal"] || natal["qimen_chart"] {
		t.Fatal("natal tool vocabulary is missing its own tools or leaks qimen tools")
	}
	if !divination["qimen_snapshot"] || divination["create_birth_chart"] {
		t.Fatal("divination tool vocabulary is missing its own tools or leaks natal tools")
	}
}

func TestSkillNameForDoc(t *testing.T) {
	got := skillNameForDoc(filepath.Join("..", "..", "..", "skills", "liki", "divination", "ENTRY.md"))
	if got != "divination" {
		t.Fatalf("skill name = %q, want divination", got)
	}
}

func collectToolVocabulary(value any, allow map[string]bool) {
	switch item := value.(type) {
	case map[string]any:
		if properties, ok := item["properties"].(map[string]any); ok {
			for name, child := range properties {
				allow[name] = true
				collectToolVocabulary(child, allow)
			}
		}
		if enums, ok := item["enum"].([]any); ok {
			for _, enum := range enums {
				if name, ok := enum.(string); ok {
					allow[name] = true
				}
			}
		}
		for key, child := range item {
			if key == "properties" {
				continue
			}
			collectToolVocabulary(child, allow)
		}
	case []any:
		for _, child := range item {
			collectToolVocabulary(child, allow)
		}
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

// fieldAllowed reports whether a field token is allowed directly or as a
// dot-separated nested path whose every segment resolves (e.g. time_scope.type).
func fieldAllowed(seg string, docAllow map[string]bool, refs *fieldRefs) bool {
	if docAllow[seg] || pathResolvable(seg, refs) {
		return true
	}
	if strings.Contains(seg, ".") {
		for _, part := range strings.Split(seg, ".") {
			if part == "" {
				continue
			}
			if !docAllow[part] && !pathResolvable(part, refs) {
				return false
			}
		}
		return true
	}
	return false
}
