package mcpcontract

// skill 文档与引擎 schema 同步测试：
//  1. liki-skills 文档（SKILL.md + references 平铺卡）反引号引用的字段名必须存在于某方法
//     Params/Result 的 JSON Schema 属性集合。
//  2. skill 侧 check_docs.py 的方法白名单 == 引擎注册方法集。
//
// 依赖本 monorepo 的 skills/ 目录（相对路径基于包目录 engine/internal/mcpcontract：../../../skills）；目录缺失时自动跳过。

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
)

// 校验范围：unified Liki skill（根入口 + 四个领域包）
// 相对路径基于包目录 engine/internal/mcpcontract（go test cwd）：../../../skills/... 指向本 monorepo 根 skills/
var skillDocsRels = []string{
	"../../../skills/liki",
}

var (
	reFieldToken = regexp.MustCompile(`^[a-zA-Z_][\w.\[\]/]*$`)
	reEnvVar     = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	reExtension  = regexp.MustCompile(`\.(md|py|json|csv|sh|yaml|yml|toml|txt)$`)
)

// protocolFieldTokens are MCP protocol schema keywords, not engine result fields.
var protocolFieldTokens = map[string]bool{
	"inputSchema":  true,
	"outputSchema": true,
}

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
					// 受控字段的取值（const/enum）也是合法的文档引用——
					// 判别联合的 type 值（如 time_scope 的 year_range/decade）
					// 与枚举字段的值（如 gender 的 male/female）都算"存在"。
					if c, ok := subMap["const"].(string); ok {
						out.leaves[c] = true
					}
					if enums, ok := subMap["enum"].([]any); ok {
						for _, e := range enums {
							if s, ok := e.(string); ok {
								out.leaves[s] = true
							}
						}
					}
					switch schemaKind(subMap) {
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
						if items, ok := subMap["items"].(map[string]any); ok {
							out.paths = append(out.paths, p+"[]")
							walk(items, p+"[].")
						} else {
							out.paths = append(out.paths, p)
						}
					}
				} else {
					out.paths = append(out.paths, p)
				}
			}
		}
	}
	// Typed DTO schemas are the published document: walk $defs (the Go
	// generator keeps them) plus the root. The retired {"_product","data"}
	// envelope is gone, so there is nothing to unwrap first.
	if obj, ok := doc.(map[string]any); ok {
		if defs, ok := obj["$defs"].(map[string]any); ok {
			for _, def := range defs {
				walk(def, "")
			}
		}
	}
	walk(doc, root)
}

func contractFieldRefs(contracts *Registry) *fieldRefs {
	out := &fieldRefs{leaves: map[string]bool{}}
	for _, name := range contracts.Names() {
		definition, err := contracts.Definition(name)
		if err != nil {
			continue
		}
		collectFieldRefs(definition.InputSchema, "params.", out)
		if len(definition.OutputSchema) > 0 {
			collectFieldRefs(definition.OutputSchema, name+".", out)
		}
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
	if refs.leaves[segs[len(segs)-1]] {
		return true
	}
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
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			// 运行时治理文档（支付 / 反馈）不属于引擎 schema 字段文档面。
			rel, relErr := filepath.Rel(filepath.Join("..", "..", "..", "skills", "liki"), path)
			if relErr == nil && (rel == filepath.Join("references", "ops-FEEDBACK.md") ||
				rel == filepath.Join("references", "payment-AIPAY.md")) {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// loadAssertionIDs 读取 counsel 断言表的 assertion_id 集合。
// skill 卡以 `cai_104` 形式引用跨层断言（counsel 数据，而非引擎 schema 字段）；
// 精确匹配可防拼写漂移，末尾 x 的通配（jk_10x）按前缀匹配。
func loadAssertionIDs() map[string]bool {
	ids := map[string]bool{}
	paths, err := filepath.Glob(
		filepath.Join("..", "..", "..", "counsel", "app", "*", "tools", "assertions", "*assertions.csv"),
	)
	if err != nil {
		return ids
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		rows, err := csv.NewReader(strings.NewReader(string(raw))).ReadAll()
		if err != nil || len(rows) < 2 {
			continue
		}
		column := -1
		for index, header := range rows[0] {
			if strings.TrimSpace(header) == "assertion_id" {
				column = index
				break
			}
		}
		if column < 0 {
			continue
		}
		for _, row := range rows[1:] {
			if column >= len(row) {
				continue
			}
			if id := strings.TrimSpace(row[column]); id != "" {
				ids[id] = true
			}
		}
	}
	return ids
}

// assertionIDAllowed 允许精确断言 id 或末尾 x 的通配前缀（jk_10x → 前缀 jk_10）。
func assertionIDAllowed(tok string, ids map[string]bool) bool {
	if ids[tok] {
		return true
	}
	prefix := strings.TrimRight(tok, "x")
	if prefix == tok {
		return false
	}
	for id := range ids {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

func TestSkillDocsFieldRefs(t *testing.T) {
	files, err := loadSkillDocs()
	if err != nil || len(files) == 0 {
		t.Skip("skills/ 目录不存在，跳过 skill 文档契约测试")
	}
	contracts := testEngine()
	refs := contractFieldRefs(contracts)
	assertionIDs := loadAssertionIDs()
	// 放行集合：引擎工具名 + 当前 skill 工具 schema 词表（上方）
	// 加上字面量列出的宽松词元——即文档中合法但不属于任何 schema 的少量散文词。
	allow := map[string]bool{}
	for _, name := range contracts.Names() {
		allow[name] = true
		// MCP 工具名 = RPC 方法名点转下划线（bazi.chart → bazi_chart）
		allow[strings.ReplaceAll(name, ".", "_")] = true
	}
	toolVocabulary, err := loadSkillToolVocabulary()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []string{
		"error", "data", "isError", "snapshot", "engine", "natal", "curl", "jian_chu", "huangdao", "gan_ji",
		"zhi_ji", "mansion", "xi_shen", "cai_shen", "suitability", "qimen_matters", "bed_install", "build",
		"cleaning", "counsel", "engage", "event_type", "exam", "funeral", "income", "medical", "move",
		"opening", "renovation", "sacrifice", "sign", "wedding",
		"finish_task", // 运行时任务提交工具（框架注入，非引擎工具）
	} {
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
				if strings.ContainsAny(tok, "{}") || reEnvVar.MatchString(tok) || !reFieldToken.MatchString(tok) { // reEnvVar: 全大写下划线=环境变量（LIKI_MCP_URL），非 schema 字段
					continue
				}
				segments := normalizeFieldToken(tok)
				protocolOnly := false
				for _, seg := range segments {
					if protocolFieldTokens[seg] {
						protocolOnly = true
						break
					}
				}
				if protocolOnly {
					continue
				}
				if reExtension.MatchString(tok) || strings.HasPrefix(tok, "tools/") ||
					strings.HasPrefix(tok, "skills/") {
					continue
				}
				// 含 '/' 但不是已知导航前缀（tools/、skills/）→ HTTP 头/媒体类型等值（如 Content-Type: application/json），非 schema 字段，跳过
				if strings.Contains(tok, "/") {
					continue
				}
				for _, seg := range normalizeFieldToken(tok) {
					if fieldAllowed(seg, docAllow, refs) || assertionIDAllowed(seg, assertionIDs) {
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

// schemaKind normalizes a schema's "type" keyword. jsonschema-go renders Go
// slices and optional objects as type unions such as ["null","array"], so a
// plain string comparison would stop the walk before nested properties.
func schemaKind(schema map[string]any) string {
	switch value := schema["type"].(type) {
	case string:
		return value
	case []any:
		for _, item := range value {
			if text, ok := item.(string); ok && text != "null" {
				return text
			}
		}
	}
	return ""
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
	if parts[1] == "references" {
		if len(parts) < 3 {
			return "liki"
		}
		return skillNameForFlatCard(parts[len(parts)-1])
	}
	return parts[1]
}

// skillNameForFlatCard 把平铺卡名映射到所属词表面：
// `references/<域>-<app|体系>-<卡>.md`；`divination-liuyao-*` / `divination-qimen-*`
// 用体系级词表，其余按域级；根共享卡（FAQ / ops- / payment-）归 liki。
func skillNameForFlatCard(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), ".md")
	if base == "FAQ" || strings.HasPrefix(base, "ops-") || strings.HasPrefix(base, "payment-") {
		return "liki"
	}
	for _, method := range []string{"liuyao", "qimen"} {
		if strings.HasPrefix(base, "divination-"+method+"-") {
			return method
		}
	}
	if idx := strings.Index(base, "-"); idx > 0 {
		return base[:idx]
	}
	return "liki"
}

// counselToolVocabulary returns the per-skill vocabulary published by the
// counsel Pydantic contracts. Those contracts are Python models, so the engine
// suite runs the repository bridge instead of reading a committed copy of the
// schemas: the models stay the single source of truth for every field a skill
// document may name.
func counselToolVocabulary() (map[string]map[string]bool, error) {
	if _, err := os.Stat(filepath.Join("..", "..", "..", "counsel", "app", "contracts")); err != nil {
		// Standalone engine checkout: the counsel tree is not part of the module.
		return nil, nil
	}
	bridge := filepath.Join("..", "..", "..", "scripts", "counsel_tool_vocabulary.py")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, counselPython(), bridge).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// The bridge runs the counsel models, so a missing dependency shows
			// up here rather than as a silently weaker document check.
			return nil, fmt.Errorf(
				"counsel tool vocabulary bridge failed (%s): %w: %s "+
					"(hint: install the counsel dependencies, or point COUNSEL_PYTHON at an interpreter that has them)",
				bridge, err, strings.TrimSpace(string(exitErr.Stderr)),
			)
		}
		return nil, fmt.Errorf("counsel tool vocabulary bridge failed (%s): %w", bridge, err)
	}
	var published map[string][]string
	if err := json.Unmarshal(output, &published); err != nil {
		return nil, fmt.Errorf("decode counsel tool vocabulary: %w", err)
	}
	vocabulary := make(map[string]map[string]bool, len(published))
	for skill, names := range published {
		entries := make(map[string]bool, len(names))
		for _, name := range names {
			entries[name] = true
		}
		vocabulary[skill] = entries
	}
	return vocabulary, nil
}

// counselPython resolves the interpreter that owns the counsel dependencies:
// COUNSEL_PYTHON when set, the venv make test-counsel bootstraps, then python3
// (the CI engine job installs the counsel dependencies into it).
func counselPython() string {
	if configured := os.Getenv("COUNSEL_PYTHON"); configured != "" {
		return configured
	}
	venv := filepath.Join("..", "..", "..", "counsel", ".venv", "bin", "python")
	if info, err := os.Stat(venv); err == nil && !info.IsDir() {
		return venv
	}
	return "python3"
}

func loadSkillToolVocabulary() (map[string]map[string]bool, error) {
	result, err := counselToolVocabulary()
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = make(map[string]map[string]bool)
	}

	// Projection, snapshot, and answer contracts are not sent to the LLM. Load
	// them only for documentation vocabulary so output fields stay externally
	// documented.
	contractPaths, err := filepath.Glob(
		filepath.Join("..", "..", "..", "counsel", "app", "*", "tools", "*contract.json"),
	)
	if err != nil {
		return nil, err
	}
	for _, contractPath := range contractPaths {
		raw, err := os.ReadFile(contractPath)
		if err != nil {
			return nil, err
		}
		var contract any
		if err := json.Unmarshal(raw, &contract); err != nil {
			return nil, fmt.Errorf("decode %s: %w", contractPath, err)
		}
		skill := counselSkill(contractPath)
		vocabulary := result[skill]
		if vocabulary == nil {
			vocabulary = make(map[string]bool)
			result[skill] = vocabulary
		}
		if document, ok := contract.(map[string]any); ok {
			if tools, ok := document["tools"].(map[string]any); ok {
				for _, schema := range tools {
					collectToolVocabulary(schema, vocabulary)
				}
				continue
			}
		}
		if strings.Contains(filepath.Base(contractPath), "projection") {
			collectProjectionVocabulary(contract, vocabulary)
		} else {
			collectToolVocabulary(contract, vocabulary)
		}
	}

	// The unified root owns the feedback contract. Keep its vocabulary under the
	// product skill so domain tool vocabularies stay scoped.
	feedbackPath := filepath.Join("..", "..", "..", "skills", "liki", "feedback.schema.json")
	if raw, err := os.ReadFile(feedbackPath); err == nil {
		var schema any
		if err := json.Unmarshal(raw, &schema); err != nil {
			return nil, err
		}
		if result["liki"] == nil {
			result["liki"] = make(map[string]bool)
		}
		collectToolVocabulary(schema, result["liki"])
	}

	// Unified divination contracts contain both domains. Domain skill docs use
	// the domain-specific vocabulary while preserving snapshot/answer fields.
	for _, domain := range []string{"liuyao", "qimen"} {
		combined := make(map[string]bool)
		maps.Copy(combined, result["divination"])
		maps.Copy(combined, result[domain])
		result[domain] = combined
	}
	return result, nil
}

// counselSkill reads the skill package from a counsel artifact path such as
// counsel/app/contracts/divination.py.
func counselSkill(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for index, part := range parts {
		if part == "app" && index+1 < len(parts) {
			return parts[index+1]
		}
	}
	return "liki"
}

func TestSkillToolVocabularyIsScoped(t *testing.T) {
	vocabulary, err := loadSkillToolVocabulary()
	if err != nil {
		t.Fatal(err)
	}
	natal := vocabulary["natal"]
	divination := vocabulary["divination"]
	if !natal["compute_factors"] || !natal["natal_query"] || !natal["period_query"] || natal["qimen_chart"] {
		t.Fatal("natal tool vocabulary is missing its own tools or leaks qimen tools")
	}
	if !divination["qimen_snapshot"] || divination["create_birth_chart"] {
		t.Fatal("divination tool vocabulary is missing its own tools or leaks natal tools")
	}
}

func TestSkillNameForDoc(t *testing.T) {
	got := skillNameForDoc(filepath.Join("..", "..", "..", "skills", "liki", "references", "natal-ENTRY.md"))
	if got != "natal" {
		t.Fatalf("skill name = %q, want natal", got)
	}
}

func TestSkillNameForFlatDivinationDomain(t *testing.T) {
	path := filepath.Join("..", "..", "..", "skills", "liki", "references", "divination-qimen-bamen.md")
	if got := skillNameForDoc(path); got != "qimen" {
		t.Fatalf("skill name = %q, want qimen", got)
	}
}

// collectToolVocabulary walks a JSON Schema collecting property names and enum
// values. It mirrors app.contracts._collect_schema_names, which serves the same
// role for the counsel models; keep the two walks in step.
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

// collectProjectionVocabulary reads the compact projection contract format.
// Unlike JSON Schema, field names are map keys under "fields" rather than keys
// of a "properties" object.
func collectProjectionVocabulary(value any, allow map[string]bool) {
	switch item := value.(type) {
	case map[string]any:
		if fields, ok := item["fields"].(map[string]any); ok {
			for name, child := range fields {
				allow[name] = true
				collectProjectionVocabulary(child, allow)
			}
		}
		for _, child := range item {
			collectProjectionVocabulary(child, allow)
		}
	case []any:
		for _, child := range item {
			collectProjectionVocabulary(child, allow)
		}
	}
}

// 双向同步：skill 侧 check_docs.py 方法白名单 == 引擎注册方法集（防两处漂移）。
func TestSkillDocsMethodListSync(t *testing.T) {
	path := filepath.Join("..", "..", "..", "scripts", "check_docs.py")
	if _, err := os.Stat(filepath.Join("..", "..", "..", "skills")); err != nil {
		t.Skip("skills/ 目录不存在，跳过 skill 文档契约测试")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		// The monorepo is present, so a missing checker is a regression, not a skip.
		t.Fatalf("%s: %v", path, err)
	}
	whitelist := map[string]bool{}
	// 只提取 METHOD_WHITELIST = {...} 块内的字符串，允许无点单名
	seg := string(raw)
	if idx := strings.Index(seg, "METHOD_WHITELIST = {"); idx >= 0 {
		seg = seg[idx:]
		if end := strings.Index(seg, "}\n"); end >= 0 {
			seg = seg[:end]
		}
	}
	for _, m := range regexp.MustCompile(`"([a-z]+(?:[._][a-z_]+)?)"`).FindAllStringSubmatch(seg, -1) {
		whitelist[m[1]] = true
	}
	regNames := testEngine().Names()
	var missingFromSkill, missingFromEngine []string
	for _, n := range regNames {
		if !whitelist[n] {
			missingFromSkill = append(missingFromSkill, n)
		}
	}
	for n := range whitelist {
		if !slices.Contains(regNames, n) {
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
