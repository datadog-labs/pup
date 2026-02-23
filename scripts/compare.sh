#!/usr/bin/env bash
# compare.sh — Compare Go (homebrew) vs Rust (debug) pup CLI output.
# Generates a self-contained HTML report at /tmp/pup-compare-<timestamp>/report.html
set -euo pipefail

# ── Configuration ────────────────────────────────────────────────────────────
GO_BIN="/opt/homebrew/bin/pup"
RUST_BIN="./target/debug/pup"
REPORT_DIR="/tmp/pup-compare-$(date +%Y%m%d-%H%M%S)"
PARALLEL_JOBS=8
RATE_LIMIT_DELAY=0.5

# All env vars that trigger agent mode
AGENT_ENV_VARS=(
  CLAUDECODE CLAUDE_CODE FORCE_AGENT_MODE
  CURSOR_AGENT CODEX OPENAI_CODEX OPENCODE AIDER CLINE
  WINDSURF_AGENT GITHUB_COPILOT AMAZON_Q AWS_Q_DEVELOPER
  GEMINI_CODE_ASSIST SRC_CODY
)

# Build the unset string for human mode
UNSET_AGENT=""
for v in "${AGENT_ENV_VARS[@]}"; do
  UNSET_AGENT="$UNSET_AGENT -u $v"
done

# Shared top-level commands (present in both Go and Rust)
SHARED_COMMANDS=(
  agent alias api-keys apm app-keys audit-logs auth cases cicd cloud
  code-coverage cost dashboards data-governance downtime error-tracking
  events hamr incidents infrastructure integrations investigations logs
  metrics misc monitors network notebooks obs-pipelines on-call
  organizations product-analytics rum scorecards security service-catalog
  slos static-analysis status-pages synthetics tags test traces usage users
  version
)

# Read-only commands safe to run with dd-auth
READ_COMMANDS=(
  "monitors list --limit 1"
  "dashboards list"
  "slos list"
  "logs search --query=* --from=5m --limit=1"
  "metrics list"
  "users list"
  "tags list"
  "misc ip-ranges"
  "misc status"
  "organizations get"
  "downtime list"
  "service-catalog list"
  "synthetics tests list"
  "rum apps list"
  "security rules list"
  "audit-logs list"
  "usage summary"
  "api-keys list"
  "notebooks list"
  "infrastructure hosts list --limit 1"
)

# ── Setup ────────────────────────────────────────────────────────────────────
mkdir -p "$REPORT_DIR"/{go,rust}

GO_VERSION=$("$GO_BIN" version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
RUST_VERSION=$("$RUST_BIN" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")

echo "=== Pup CLI Comparison ==="
echo "Go:   $GO_BIN (v$GO_VERSION)"
echo "Rust: $RUST_BIN (v$RUST_VERSION)"
echo "Output: $REPORT_DIR/report.html"
echo ""

# Results file: tab-separated lines of  test_id  status  command  env_context  go_file  rust_file
RESULTS_TSV="$REPORT_DIR/_results.tsv"
: > "$RESULTS_TSV"

record_result() {
  local test_id="$1" status="$2" cmd="$3" env_ctx="$4" go_file="${5:-}" rust_file="${6:-}"
  printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$test_id" "$status" "$cmd" "$env_ctx" "$go_file" "$rust_file" >> "$RESULTS_TSV"
}

# ── Embedded Python helper ───────────────────────────────────────────────────
COMPARE_PY="$REPORT_DIR/_compare.py"
cat > "$COMPARE_PY" << 'PYEOF'
#!/usr/bin/env python3
"""JSON comparison and HTML report generation for pup CLI comparison."""
import json, sys, re, os, html

# ── Normalization ────────────────────────────────────────────────────────────

def normalize_json(text, go_ver="", rust_ver=""):
    text = text.strip()
    if not text:
        return text
    if go_ver:
        text = text.replace(go_ver, "VERSION")
    if rust_ver:
        text = text.replace(rust_ver, "VERSION")
    text = text.replace("\\u003c", "<").replace("\\u003e", ">").replace("\\u0026", "&")
    try:
        obj = json.loads(text)
        obj = _normalize_value(obj)
        return json.dumps(obj, indent=2, sort_keys=True, ensure_ascii=False)
    except json.JSONDecodeError:
        return text

def _normalize_value(v):
    if isinstance(v, dict):
        return {k: _normalize_value(val) for k, val in sorted(v.items())}
    elif isinstance(v, list):
        return [_normalize_value(item) for item in v]
    elif isinstance(v, str):
        v = re.sub(r'\b\d+\.\d+\.\d+\b', 'VERSION', v)
        # Normalize integer type strings: int32, int64 -> int
        v = re.sub(r'\bint(?:32|64)\b', 'int', v)
        return v
    elif isinstance(v, float) and v == int(v):
        return int(v)
    return v

def extract_schema(text):
    text = text.strip()
    if not text:
        return "{}"
    try:
        obj = json.loads(text)
        schema = _extract_schema_value(obj)
        return json.dumps(schema, indent=2, sort_keys=True)
    except json.JSONDecodeError:
        return "{}"

def _extract_schema_value(v, depth=0):
    if depth > 10:
        return type(v).__name__
    if isinstance(v, dict):
        return {k: _extract_schema_value(val, depth+1) for k, val in sorted(v.items())}
    elif isinstance(v, list):
        return "[]" if len(v) == 0 else [_extract_schema_value(v[0], depth+1)]
    elif isinstance(v, bool):
        return "bool"
    elif isinstance(v, int):
        return "int"
    elif isinstance(v, float):
        return "number"
    elif isinstance(v, str):
        return "string"
    elif v is None:
        return "null"
    return type(v).__name__

# ── Comparison ───────────────────────────────────────────────────────────────

def compare_agent_json(go_file, rust_file, go_ver, rust_ver):
    try:
        go_text = open(go_file).read()
        rust_text = open(rust_file).read()
    except FileNotFoundError as e:
        return "error"
    go_norm = normalize_json(go_text, go_ver, rust_ver)
    rust_norm = normalize_json(rust_text, go_ver, rust_ver)
    return "pass" if go_norm == rust_norm else "diff"

def compare_json_structure(go_file, rust_file):
    try:
        go_text = open(go_file).read()
        rust_text = open(rust_file).read()
    except FileNotFoundError:
        return "error"
    return "pass" if extract_schema(go_text) == extract_schema(rust_text) else "diff"

# ── HTML Report Generation ───────────────────────────────────────────────────

SENSITIVE_VARS = {"DD_API_KEY", "DD_APP_KEY", "DD_APPLICATION_KEY", "DD_ACCESS_TOKEN"}

def mask_env(env_str):
    """Mask sensitive env var values, showing only first 4 chars."""
    parts = []
    for item in env_str.split():
        if "=" in item:
            key, val = item.split("=", 1)
            if key in SENSITIVE_VARS and len(val) > 4:
                parts.append(f"{key}={val[:4]}{'*' * (len(val) - 4)}")
            else:
                parts.append(item)
        else:
            parts.append(item)
    return " ".join(parts)

def generate_diff_lines(go_text, rust_text):
    """Generate unified diff lines between two texts."""
    import difflib
    go_lines = go_text.splitlines(keepends=True)
    rust_lines = rust_text.splitlines(keepends=True)
    return list(difflib.unified_diff(go_lines, rust_lines,
                                      fromfile="Go", tofile="Rust", lineterm=""))

def render_unified_diff(go_text, rust_text):
    """Render unified diff as HTML."""
    lines = generate_diff_lines(go_text, rust_text)
    if not lines:
        return '<span class="diff-ctx">No differences</span>'
    parts = []
    for line in lines:
        line = line.rstrip('\n')
        escaped = html.escape(line)
        if line.startswith('---') or line.startswith('+++'):
            parts.append(f'<span class="diff-hdr">{escaped}</span>')
        elif line.startswith('@@'):
            parts.append(f'<span class="diff-hdr">{escaped}</span>')
        elif line.startswith('+'):
            parts.append(f'<span class="diff-add">{escaped}</span>')
        elif line.startswith('-'):
            parts.append(f'<span class="diff-del">{escaped}</span>')
        else:
            parts.append(f'<span class="diff-ctx">{escaped}</span>')
    return '\n'.join(parts)

def render_side_by_side(go_text, rust_text):
    """Render side-by-side comparison as HTML table."""
    import difflib
    go_lines = go_text.splitlines()
    rust_lines = rust_text.splitlines()
    sm = difflib.SequenceMatcher(None, go_lines, rust_lines)
    rows = []
    for op, i1, i2, j1, j2 in sm.get_opcodes():
        if op == 'equal':
            for i in range(i1, i2):
                rows.append(('ctx', go_lines[i], rust_lines[j1 + (i - i1)]))
        elif op == 'replace':
            max_len = max(i2 - i1, j2 - j1)
            for k in range(max_len):
                left = html.escape(go_lines[i1 + k]) if (i1 + k) < i2 else ''
                right = html.escape(rust_lines[j1 + k]) if (j1 + k) < j2 else ''
                left_cls = 'sbs-del' if (i1 + k) < i2 else 'sbs-empty'
                right_cls = 'sbs-add' if (j1 + k) < j2 else 'sbs-empty'
                rows.append(('change', (left_cls, left), (right_cls, right)))
        elif op == 'delete':
            for i in range(i1, i2):
                rows.append(('change', ('sbs-del', html.escape(go_lines[i])), ('sbs-empty', '')))
        elif op == 'insert':
            for j in range(j1, j2):
                rows.append(('change', ('sbs-empty', ''), ('sbs-add', html.escape(rust_lines[j]))))

    html_rows = []
    for row in rows:
        if row[0] == 'ctx':
            escaped = html.escape(row[1])
            html_rows.append(f'<tr><td class="sbs-ctx">{escaped}</td><td class="sbs-ctx">{escaped}</td></tr>')
        else:
            _, (lcls, left), (rcls, right) = row
            html_rows.append(f'<tr><td class="{lcls}">{left}</td><td class="{rcls}">{right}</td></tr>')

    return '<table class="sbs-table"><thead><tr><th>Go</th><th>Rust</th></tr></thead><tbody>' + \
           '\n'.join(html_rows) + '</tbody></table>'

def extract_top_command(test_id):
    """Extract the top-level command name from a test ID.
    HELP-AGENT-logs -> logs, CMD-monitors-list-limit-1 -> monitors,
    HELP-AGENT-toplevel -> toplevel, AUTH-misc-status -> misc"""
    # Strip category prefix
    for prefix in ('HELP-AGENT-', 'HELP-HUMAN-', 'CMD-', 'AUTH-', 'FMT-'):
        if test_id.startswith(prefix):
            rest = test_id[len(prefix):]
            # Return first segment (the top-level command)
            return rest.split('-')[0] if rest else 'toplevel'
    return 'other'

MAX_DISPLAY_LINES = 500  # Truncate very large outputs for display

def truncate_text(text, max_lines=MAX_DISPLAY_LINES):
    """Truncate text to max_lines, adding a notice if truncated."""
    lines = text.splitlines()
    if len(lines) <= max_lines:
        return text
    return '\n'.join(lines[:max_lines]) + f'\n\n... ({len(lines) - max_lines} more lines truncated)'

def render_test_row(t, go_ver, rust_ver, out, rendered_ids=None):
    """Render a single test row with its detail panel."""
    if rendered_ids is None:
        rendered_ids = set()

    cat_prefix = t['id'].split('-')[0]
    if t['id'].startswith('HELP-AGENT'):
        cat_prefix = 'HELP-AGENT'
    elif t['id'].startswith('HELP-HUMAN'):
        cat_prefix = 'HELP-HUMAN'

    row_id = re.sub(r'[^a-z0-9]', '-', t['id'].lower())
    escaped_cmd = html.escape(t['cmd'])
    masked_env = html.escape(mask_env(t['env']))

    out.write(f'''
    <div class="test-row" data-status="{t['status']}" onclick="toggleDiff('{row_id}')">
      <span class="badge {t['status']}">{t['status']}</span>
      <span class="test-id">{html.escape(t['id'])}</span>
      <span class="test-cmd">{escaped_cmd}</span>
    </div>
''')

    # Only render detail panels for diff/error, and only once per test ID
    if t['status'] not in ('diff', 'error'):
        return
    if row_id in rendered_ids:
        return
    rendered_ids.add(row_id)

    env_line = '<div class="env-ctx">'
    if t['env']:
        env_line += f'<span class="env-label">ENV:</span> <code>{masked_env}</code> '
    env_line += f'<span class="env-label">CMD:</span> <code>{escaped_cmd}</code></div>'

    if t['go_file'] and t['rust_file'] and os.path.isfile(t['go_file']) and os.path.isfile(t['rust_file']):
        go_text = open(t['go_file']).read()
        rust_text = open(t['rust_file']).read()

        is_json_cat = cat_prefix in ('HELP-AGENT', 'CMD', 'FMT')
        # For CMD/FMT with large output, compare schemas not full text
        is_api_output = cat_prefix in ('CMD', 'FMT')
        if is_api_output:
            go_display = truncate_text(extract_schema(go_text))
            rust_display = truncate_text(extract_schema(rust_text))
        elif is_json_cat:
            go_display = truncate_text(normalize_json(go_text, go_ver, rust_ver))
            rust_display = truncate_text(normalize_json(rust_text, go_ver, rust_ver))
        else:
            go_display = truncate_text(go_text)
            rust_display = truncate_text(rust_text)

        unified_html = render_unified_diff(go_display, rust_display)
        sbs_html = render_side_by_side(go_display, rust_display)

        out.write(f'''
    <div class="detail-panel" id="diff-{row_id}">
      {env_line}
      <div class="view-tabs">
        <button class="vtab active" onclick="switchView(this, 'unified')">Unified Diff</button>
        <button class="vtab" onclick="switchView(this, 'sidebyside')">Side by Side</button>
      </div>
      <div class="view-unified">{unified_html}</div>
      <div class="view-sidebyside" style="display:none">{sbs_html}</div>
    </div>
''')
    else:
        out.write(f'''
    <div class="detail-panel" id="diff-{row_id}">
      {env_line}
      <div class="diff-ctx">No output files available for comparison</div>
    </div>
''')

def count_statuses(test_list):
    cc = {}
    for t in test_list:
        cc[t['status']] = cc.get(t['status'], 0) + 1
    return cc

def stats_str(cc):
    return f"{cc.get('pass',0)} pass, {cc.get('diff',0)} diff, {cc.get('error',0)} error, {cc.get('skip',0)} skip"

def generate_report(results_tsv, report_dir, go_bin, rust_bin, go_ver, rust_ver):
    """Generate HTML report from results TSV and raw output files."""
    tests = []
    with open(results_tsv) as f:
        for line in f:
            line = line.rstrip('\n')
            if not line:
                continue
            parts = line.split('\t')
            tests.append({
                'id': parts[0],
                'status': parts[1],
                'cmd': parts[2],
                'env': parts[3],
                'go_file': parts[4] if len(parts) > 4 else '',
                'rust_file': parts[5] if len(parts) > 5 else '',
            })

    total = len(tests)
    counts = count_statuses(tests)

    categories = [
        ('MISSING', 'Missing Commands'),
        ('HELP-AGENT', 'Agent-Mode Help'),
        ('HELP-HUMAN', 'Human-Mode Help'),
        ('CMD', 'Read-Only Commands'),
        ('AUTH', 'Auth Checks'),
        ('FMT', 'Output Format Parity'),
    ]

    # Build command groups: command_name -> list of tests
    from collections import OrderedDict
    cmd_groups = OrderedDict()
    for t in sorted(tests, key=lambda x: x['id']):
        cmd_name = extract_top_command(t['id'])
        cmd_groups.setdefault(cmd_name, []).append(t)

    rendered_ids = set()  # Track which detail panels have been rendered

    report_path = os.path.join(report_dir, 'report.html')
    with open(report_path, 'w') as out:
        out.write(HTML_HEAD)
        out.write(f'''
<h1>Pup CLI Comparison Report</h1>
<div class="meta">
  Go: <strong>{html.escape(go_bin)}</strong> v{html.escape(go_ver)} &nbsp;|&nbsp;
  Rust: <strong>{html.escape(rust_bin)}</strong> v{html.escape(rust_ver)} &nbsp;|&nbsp;
  Generated: {__import__("datetime").datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
</div>
<div class="legend">
  <span class="legend-item"><span class="swatch swatch-del"></span> Go only (removed in Rust)</span>
  <span class="legend-item"><span class="swatch swatch-add"></span> Rust only (added vs Go)</span>
</div>
<div class="summary">
  <div class="card total"><span class="num">{total}</span><span class="label">Total</span></div>
  <div class="card pass"><span class="num">{counts.get("pass",0)}</span><span class="label">Pass</span></div>
  <div class="card diff"><span class="num">{counts.get("diff",0)}</span><span class="label">Diff</span></div>
  <div class="card error"><span class="num">{counts.get("error",0)}</span><span class="label">Error</span></div>
  <div class="card skip"><span class="num">{counts.get("skip",0)}</span><span class="label">Skip</span></div>
</div>
<div class="controls">
  <div class="filters">
    <button class="active" onclick="filterTests('all')">All</button>
    <button onclick="filterTests('pass')">Pass</button>
    <button onclick="filterTests('diff')">Diff</button>
    <button onclick="filterTests('error')">Error</button>
    <button onclick="filterTests('skip')">Skip</button>
  </div>
  <div class="group-toggle">
    <span class="group-label">Group by:</span>
    <button class="gtab active" onclick="switchGroup('category')">Category</button>
    <button class="gtab" onclick="switchGroup('command')">Command</button>
  </div>
</div>
''')

        # ── Category view ──
        out.write('<div id="view-category">\n')
        for cat_prefix, cat_name in categories:
            cat_tests = [t for t in tests if t['id'].startswith(cat_prefix + '-')]
            if not cat_tests:
                continue
            cat_tests.sort(key=lambda t: t['id'])
            cc = count_statuses(cat_tests)

            out.write(f'''
<div class="category">
  <h2 onclick="this.parentElement.classList.toggle('collapsed')">
    {cat_name} <span class="cat-stats">({stats_str(cc)})</span>
  </h2>
  <div class="tests">
''')
            for t in cat_tests:
                render_test_row(t, go_ver, rust_ver, out, rendered_ids)
            out.write('  </div>\n</div>\n')

        out.write('</div>\n')

        # ── Command view ──
        out.write('<div id="view-command" style="display:none">\n')
        sorted_cmds = sorted(cmd_groups.keys(), key=lambda x: ('0' if x == 'toplevel' else '1' + x))
        for cmd_name in sorted_cmds:
            group_tests = cmd_groups[cmd_name]
            cc = count_statuses(group_tests)
            display_name = 'Top-Level' if cmd_name == 'toplevel' else cmd_name

            out.write(f'''
<div class="category">
  <h2 onclick="this.parentElement.classList.toggle('collapsed')">
    {html.escape(display_name)} <span class="cat-stats">({stats_str(cc)})</span>
  </h2>
  <div class="tests">
''')
            for t in group_tests:
                render_test_row(t, go_ver, rust_ver, out, rendered_ids)
            out.write('  </div>\n</div>\n')

        out.write('</div>\n')

        out.write(HTML_FOOT)
    return report_path

# ── HTML Templates ───────────────────────────────────────────────────────────

HTML_HEAD = '''<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Pup CLI Comparison Report</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
         background: #0d1117; color: #c9d1d9; padding: 20px; line-height: 1.5; }
  h1 { color: #58a6ff; margin-bottom: 4px; }
  .meta { color: #8b949e; margin-bottom: 12px; font-size: 14px; }
  .legend { margin-bottom: 16px; display: flex; gap: 20px; font-size: 13px; color: #8b949e; }
  .legend-item { display: flex; align-items: center; gap: 6px; }
  .swatch { display: inline-block; width: 14px; height: 14px; border-radius: 3px; }
  .swatch-del { background: #3d1214; border: 1px solid #f85149; }
  .swatch-add { background: #0d2818; border: 1px solid #3fb950; }
  .summary { display: flex; gap: 12px; margin-bottom: 24px; flex-wrap: wrap; }
  .summary .card { padding: 12px 20px; border-radius: 8px; background: #161b22;
                   border: 1px solid #30363d; min-width: 100px; text-align: center; }
  .card .num { font-size: 28px; font-weight: bold; display: block; }
  .card .label { font-size: 12px; color: #8b949e; text-transform: uppercase; }
  .card.pass .num { color: #3fb950; }
  .card.diff .num { color: #d29922; }
  .card.error .num { color: #f85149; }
  .card.skip .num { color: #8b949e; }
  .card.total .num { color: #58a6ff; }
  .controls { display: flex; justify-content: space-between; align-items: center;
              margin-bottom: 16px; flex-wrap: wrap; gap: 12px; }
  .filters { display: flex; gap: 8px; flex-wrap: wrap; }
  .filters button { padding: 6px 16px; border-radius: 20px; border: 1px solid #30363d;
                    background: #161b22; color: #c9d1d9; cursor: pointer; font-size: 13px; }
  .filters button:hover { border-color: #58a6ff; }
  .filters button.active { background: #1f6feb; border-color: #1f6feb; color: #fff; }
  .group-toggle { display: flex; align-items: center; gap: 6px; }
  .group-label { color: #8b949e; font-size: 13px; }
  .gtab { padding: 4px 12px; border-radius: 6px; border: 1px solid #30363d;
          background: #161b22; color: #8b949e; cursor: pointer; font-size: 12px; }
  .gtab:hover { border-color: #58a6ff; }
  .gtab.active { background: #1f6feb; border-color: #1f6feb; color: #fff; }
  .category { margin-bottom: 24px; }
  .category h2 { color: #58a6ff; font-size: 18px; margin-bottom: 8px;
                  cursor: pointer; user-select: none; }
  .category h2::before { content: "\\25BC "; font-size: 12px; }
  .category.collapsed h2::before { content: "\\25B6 "; }
  .category.collapsed .tests { display: none; }
  .cat-stats { font-size: 13px; color: #8b949e; font-weight: normal; }
  .test-row { display: flex; align-items: center; gap: 10px; padding: 8px 12px;
              border-bottom: 1px solid #21262d; cursor: pointer; font-size: 13px; }
  .test-row:hover { background: #161b22; }
  .test-row.hidden { display: none; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 10px;
           font-size: 11px; font-weight: 600; min-width: 48px; text-align: center; flex-shrink: 0; }
  .badge.pass { background: #0d2818; color: #3fb950; }
  .badge.diff { background: #2a1f00; color: #d29922; }
  .badge.error { background: #2d0000; color: #f85149; }
  .badge.skip { background: #1c1c1c; color: #8b949e; }
  .test-id { font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace; color: #79c0ff;
             min-width: 240px; font-size: 12px; flex-shrink: 0; }
  .test-cmd { color: #8b949e; font-size: 12px; flex: 1; overflow: hidden;
              text-overflow: ellipsis; white-space: nowrap; }

  /* Detail panel */
  .detail-panel { display: none; background: #161b22; border: 1px solid #30363d;
                  border-radius: 6px; margin: 4px 0 8px 0; padding: 16px;
                  max-height: 600px; overflow: auto; }
  .detail-panel.open { display: block; }
  .env-ctx { font-size: 12px; color: #8b949e; margin-bottom: 12px; padding: 8px 10px;
             background: #0d1117; border-radius: 4px; font-family: "SF Mono", monospace; }
  .env-label { color: #d29922; font-weight: 600; }
  .env-ctx code { color: #c9d1d9; }

  /* View tabs */
  .view-tabs { margin-bottom: 10px; display: flex; gap: 4px; }
  .vtab { padding: 4px 12px; border-radius: 6px 6px 0 0; border: 1px solid #30363d;
          border-bottom: none; background: #0d1117; color: #8b949e; cursor: pointer; font-size: 12px; }
  .vtab.active { background: #161b22; color: #c9d1d9; border-bottom: 1px solid #161b22;
                 margin-bottom: -1px; position: relative; z-index: 1; }

  /* Unified diff */
  .view-unified { font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace; font-size: 12px;
                  white-space: pre-wrap; word-break: break-all; line-height: 1.5;
                  border: 1px solid #30363d; border-radius: 4px; padding: 8px;
                  background: #0d1117; }
  .diff-add { color: #3fb950; background: #0d2818; display: block; }
  .diff-del { color: #f85149; background: #3d1214; display: block; }
  .diff-hdr { color: #79c0ff; display: block; }
  .diff-ctx { color: #8b949e; display: block; }

  /* Side-by-side */
  .view-sidebyside { border: 1px solid #30363d; border-radius: 4px; overflow: auto; }
  .sbs-table { width: 100%; border-collapse: collapse;
               font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace; font-size: 12px; }
  .sbs-table th { background: #21262d; color: #8b949e; padding: 6px 10px; text-align: left;
                  font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px;
                  position: sticky; top: 0; border-bottom: 2px solid #30363d; }
  .sbs-table td { padding: 2px 8px; vertical-align: top; white-space: pre-wrap;
                  word-break: break-all; width: 50%; border-right: 1px solid #21262d; }
  .sbs-ctx { color: #8b949e; background: #0d1117; }
  .sbs-del { color: #f85149; background: #3d1214; }
  .sbs-add { color: #3fb950; background: #0d2818; }
  .sbs-empty { background: #161b22; color: #30363d; }
</style>
</head>
<body>
'''

HTML_FOOT = '''
<script>
function filterTests(status) {
  document.querySelectorAll('.filters button').forEach(b => b.classList.remove('active'));
  event.target.classList.add('active');
  // Apply to whichever view is visible
  document.querySelectorAll('.test-row').forEach(row => {
    if (status === 'all' || row.dataset.status === status) {
      row.classList.remove('hidden');
    } else {
      row.classList.add('hidden');
    }
  });
  document.querySelectorAll('.detail-panel').forEach(p => p.classList.remove('open'));
}
function toggleDiff(id) {
  const panel = document.getElementById('diff-' + id);
  if (panel) panel.classList.toggle('open');
}
function switchView(btn, view) {
  const panel = btn.closest('.detail-panel');
  panel.querySelectorAll('.vtab').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  panel.querySelector('.view-unified').style.display = (view === 'unified') ? '' : 'none';
  panel.querySelector('.view-sidebyside').style.display = (view === 'sidebyside') ? '' : 'none';
}
function switchGroup(mode) {
  document.querySelectorAll('.gtab').forEach(b => b.classList.remove('active'));
  event.target.classList.add('active');
  document.getElementById('view-category').style.display = (mode === 'category') ? '' : 'none';
  document.getElementById('view-command').style.display = (mode === 'command') ? '' : 'none';
  // Re-apply current filter
  const activeFilter = document.querySelector('.filters button.active');
  if (activeFilter) activeFilter.click();
}
</script>
</body>
</html>
'''

# ── CLI ──────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    cmd = sys.argv[1]
    if cmd == "normalize":
        text = sys.stdin.read()
        go_ver = sys.argv[2] if len(sys.argv) > 2 else ""
        rust_ver = sys.argv[3] if len(sys.argv) > 3 else ""
        print(normalize_json(text, go_ver, rust_ver))
    elif cmd == "compare-agent":
        go_file, rust_file = sys.argv[2], sys.argv[3]
        go_ver = sys.argv[4] if len(sys.argv) > 4 else ""
        rust_ver = sys.argv[5] if len(sys.argv) > 5 else ""
        print(compare_agent_json(go_file, rust_file, go_ver, rust_ver))
    elif cmd == "compare-structure":
        go_file, rust_file = sys.argv[2], sys.argv[3]
        print(compare_json_structure(go_file, rust_file))
    elif cmd == "schema":
        text = sys.stdin.read()
        print(extract_schema(text))
    elif cmd == "report":
        results_tsv = sys.argv[2]
        report_dir = sys.argv[3]
        go_bin = sys.argv[4]
        rust_bin = sys.argv[5]
        go_ver = sys.argv[6]
        rust_ver = sys.argv[7]
        path = generate_report(results_tsv, report_dir, go_bin, rust_bin, go_ver, rust_ver)
        print(path)
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        sys.exit(1)
PYEOF
chmod +x "$COMPARE_PY"

# ── Category 1: Agent-mode help ──────────────────────────────────────────────
run_agent_help_tests() {
  echo "Category 1: Agent-mode help tests..."
  local pids=()

  # Top-level help
  (
    env FORCE_AGENT_MODE=1 "$GO_BIN" --help > "$REPORT_DIR/go/agent-help-toplevel.json" 2>&1 || true
    env FORCE_AGENT_MODE=1 "$RUST_BIN" --help > "$REPORT_DIR/rust/agent-help-toplevel.json" 2>&1 || true
  ) &
  pids+=($!)

  for cmd in "${SHARED_COMMANDS[@]}"; do
    (
      env FORCE_AGENT_MODE=1 "$GO_BIN" "$cmd" --help > "$REPORT_DIR/go/agent-help-${cmd}.json" 2>&1 || true
      env FORCE_AGENT_MODE=1 "$RUST_BIN" "$cmd" --help > "$REPORT_DIR/rust/agent-help-${cmd}.json" 2>&1 || true
    ) &
    pids+=($!)
    if (( ${#pids[@]} >= PARALLEL_JOBS )); then
      wait "${pids[0]}" 2>/dev/null || true
      pids=("${pids[@]:1}")
    fi
  done
  for pid in "${pids[@]}"; do wait "$pid" 2>/dev/null || true; done

  # Compare
  local env_ctx="FORCE_AGENT_MODE=1"

  local tid="HELP-AGENT-toplevel"
  local go_f="$REPORT_DIR/go/agent-help-toplevel.json"
  local rust_f="$REPORT_DIR/rust/agent-help-toplevel.json"
  if [[ -f "$go_f" && -f "$rust_f" ]]; then
    local result
    result=$(python3 "$COMPARE_PY" compare-agent "$go_f" "$rust_f" "$GO_VERSION" "$RUST_VERSION")
    record_result "$tid" "$result" "pup --help" "$env_ctx" "$go_f" "$rust_f"
  else
    record_result "$tid" "error" "pup --help" "$env_ctx" "$go_f" "$rust_f"
  fi

  for cmd in "${SHARED_COMMANDS[@]}"; do
    tid="HELP-AGENT-${cmd}"
    go_f="$REPORT_DIR/go/agent-help-${cmd}.json"
    rust_f="$REPORT_DIR/rust/agent-help-${cmd}.json"
    if [[ -f "$go_f" && -f "$rust_f" ]]; then
      local result
      result=$(python3 "$COMPARE_PY" compare-agent "$go_f" "$rust_f" "$GO_VERSION" "$RUST_VERSION")
      record_result "$tid" "$result" "pup $cmd --help" "$env_ctx" "$go_f" "$rust_f"
    else
      record_result "$tid" "error" "pup $cmd --help" "$env_ctx" "$go_f" "$rust_f"
    fi
  done

  echo "  Done: $((${#SHARED_COMMANDS[@]}+1)) tests"
}

# ── Category 2: Human-mode help ─────────────────────────────────────────────
run_human_help_tests() {
  echo "Category 2: Human-mode help tests..."
  local pids=()
  local env_ctx="(all agent env vars unset)"

  (
    eval "env $UNSET_AGENT" "$GO_BIN" --help > "$REPORT_DIR/go/human-help-toplevel.txt" 2>&1 || true
    eval "env $UNSET_AGENT" "$RUST_BIN" --help > "$REPORT_DIR/rust/human-help-toplevel.txt" 2>&1 || true
  ) &
  pids+=($!)

  for cmd in "${SHARED_COMMANDS[@]}"; do
    (
      eval "env $UNSET_AGENT" "$GO_BIN" "$cmd" --help > "$REPORT_DIR/go/human-help-${cmd}.txt" 2>&1 || true
      eval "env $UNSET_AGENT" "$RUST_BIN" "$cmd" --help > "$REPORT_DIR/rust/human-help-${cmd}.txt" 2>&1 || true
    ) &
    pids+=($!)
    if (( ${#pids[@]} >= PARALLEL_JOBS )); then
      wait "${pids[0]}" 2>/dev/null || true
      pids=("${pids[@]:1}")
    fi
  done
  for pid in "${pids[@]}"; do wait "$pid" 2>/dev/null || true; done

  # Compare
  local tid go_f rust_f
  tid="HELP-HUMAN-toplevel"
  go_f="$REPORT_DIR/go/human-help-toplevel.txt"
  rust_f="$REPORT_DIR/rust/human-help-toplevel.txt"
  if [[ -f "$go_f" && -f "$rust_f" ]]; then
    if diff -q "$go_f" "$rust_f" &>/dev/null; then
      record_result "$tid" "pass" "pup --help" "$env_ctx" "$go_f" "$rust_f"
    else
      record_result "$tid" "diff" "pup --help" "$env_ctx" "$go_f" "$rust_f"
    fi
  else
    record_result "$tid" "error" "pup --help" "$env_ctx" "$go_f" "$rust_f"
  fi

  for cmd in "${SHARED_COMMANDS[@]}"; do
    tid="HELP-HUMAN-${cmd}"
    go_f="$REPORT_DIR/go/human-help-${cmd}.txt"
    rust_f="$REPORT_DIR/rust/human-help-${cmd}.txt"
    if [[ -f "$go_f" && -f "$rust_f" ]]; then
      if diff -q "$go_f" "$rust_f" &>/dev/null; then
        record_result "$tid" "pass" "pup $cmd --help" "$env_ctx" "$go_f" "$rust_f"
      else
        record_result "$tid" "diff" "pup $cmd --help" "$env_ctx" "$go_f" "$rust_f"
      fi
    else
      record_result "$tid" "error" "pup $cmd --help" "$env_ctx" "$go_f" "$rust_f"
    fi
  done

  echo "  Done: $((${#SHARED_COMMANDS[@]}+1)) tests"
}

# ── Category 3: Read-only commands via dd-auth ───────────────────────────────
run_read_tests() {
  echo "Category 3: Read-only commands via dd-auth..."
  local cmd_str tid slug

  # Capture env from dd-auth for display (masked)
  local dd_env
  dd_env=$(dd-auth -- env 2>/dev/null | grep -E '^DD_' | head -10 || echo "")
  local env_ctx
  env_ctx=$(echo "$dd_env" | tr '\n' ' ')

  for cmd_str in "${READ_COMMANDS[@]}"; do
    slug="${cmd_str// /-}"
    slug="${slug//=/-}"
    slug="${slug//\*/star}"
    slug="${slug//--/}"
    tid="CMD-${slug}"

    echo "  Running: pup $cmd_str"

    eval "dd-auth -- $GO_BIN $cmd_str" > "$REPORT_DIR/go/cmd-${slug}.json" 2>&1 || true
    sleep "$RATE_LIMIT_DELAY"
    eval "dd-auth -- $RUST_BIN $cmd_str" > "$REPORT_DIR/rust/cmd-${slug}.json" 2>&1 || true
    sleep "$RATE_LIMIT_DELAY"

    local go_f="$REPORT_DIR/go/cmd-${slug}.json"
    local rust_f="$REPORT_DIR/rust/cmd-${slug}.json"

    if [[ -f "$go_f" && -f "$rust_f" ]]; then
      local go_size rust_size
      go_size=$(wc -c < "$go_f" | tr -d ' ')
      rust_size=$(wc -c < "$rust_f" | tr -d ' ')

      if (( go_size < 10 )) || (( rust_size < 10 )); then
        record_result "$tid" "error" "dd-auth -- pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
        continue
      fi

      local result
      result=$(python3 "$COMPARE_PY" compare-structure "$go_f" "$rust_f")
      record_result "$tid" "$result" "dd-auth -- pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
    else
      record_result "$tid" "error" "dd-auth -- pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
    fi
  done

  echo "  Done: ${#READ_COMMANDS[@]} tests"
}

# ── Category 4: Auth requirement checks ─────────────────────────────────────
run_auth_tests() {
  echo "Category 4: Auth requirement checks..."
  local pids=()
  local env_ctx="DD_API_KEY=(unset) DD_APP_KEY=(unset) DD_ACCESS_TOKEN=(unset) DD_SITE=noauth.invalid"
  local auth_cmds=(
    "monitors list --limit 1"
    "dashboards list"
    "slos list"
    "logs search --query=* --from=5m --limit=1"
    "metrics list"
    "users list"
    "tags list"
    "misc ip-ranges"
    "misc status"
    "organizations get"
    "downtime list"
    "synthetics tests list"
    "rum apps list"
    "security rules list"
    "audit-logs list"
    "usage summary"
    "api-keys list"
    "notebooks list"
  )

  for cmd_str in "${auth_cmds[@]}"; do
    local slug="${cmd_str// /-}"
    slug="${slug//=/-}"
    slug="${slug//\*/star}"
    slug="${slug//--/}"
    local tid="AUTH-${slug}"

    (
      env -u DD_API_KEY -u DD_APP_KEY -u DD_ACCESS_TOKEN DD_SITE=noauth.invalid \
        "$GO_BIN" $cmd_str > "$REPORT_DIR/go/auth-${slug}.txt" 2>&1 || true
      env -u DD_API_KEY -u DD_APP_KEY -u DD_ACCESS_TOKEN DD_SITE=noauth.invalid \
        "$RUST_BIN" $cmd_str > "$REPORT_DIR/rust/auth-${slug}.txt" 2>&1 || true
    ) &
    pids+=($!)

    if (( ${#pids[@]} >= PARALLEL_JOBS )); then
      wait "${pids[0]}" 2>/dev/null || true
      pids=("${pids[@]:1}")
    fi
  done
  for pid in "${pids[@]}"; do wait "$pid" 2>/dev/null || true; done

  for cmd_str in "${auth_cmds[@]}"; do
    local slug="${cmd_str// /-}"
    slug="${slug//=/-}"
    slug="${slug//\*/star}"
    slug="${slug//--/}"
    local tid="AUTH-${slug}"

    local go_f="$REPORT_DIR/go/auth-${slug}.txt"
    local rust_f="$REPORT_DIR/rust/auth-${slug}.txt"

    if [[ -f "$go_f" && -f "$rust_f" ]]; then
      local go_ok=0 rust_ok=0
      if grep -q '"status".*"success"' "$go_f" 2>/dev/null; then go_ok=1; fi
      if grep -q '"status".*"success"' "$rust_f" 2>/dev/null; then rust_ok=1; fi

      if [[ "$go_ok" -eq "$rust_ok" ]]; then
        record_result "$tid" "pass" "pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
      else
        record_result "$tid" "diff" "pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
      fi
    else
      record_result "$tid" "error" "pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
    fi
  done

  echo "  Done: ${#auth_cmds[@]} tests"
}

# ── Category 5: Output format parity ────────────────────────────────────────
run_format_tests() {
  echo "Category 5: Output format parity..."
  local cmd_str slug tid

  local dd_env
  dd_env=$(dd-auth -- env 2>/dev/null | grep -E '^DD_' | head -10 || echo "")
  local env_ctx
  env_ctx=$(echo "$dd_env" | tr '\n' ' ')

  for cmd_str in "${READ_COMMANDS[@]}"; do
    slug="${cmd_str// /-}"
    slug="${slug//=/-}"
    slug="${slug//\*/star}"
    slug="${slug//--/}"
    tid="FMT-${slug}"

    local go_f="$REPORT_DIR/go/cmd-${slug}.json"
    local rust_f="$REPORT_DIR/rust/cmd-${slug}.json"

    if [[ ! -f "$go_f" || ! -f "$rust_f" ]]; then
      record_result "$tid" "skip" "pup $cmd_str" "$env_ctx"
      continue
    fi

    local go_size rust_size
    go_size=$(wc -c < "$go_f" | tr -d ' ')
    rust_size=$(wc -c < "$rust_f" | tr -d ' ')

    if (( go_size < 10 )) || (( rust_size < 10 )); then
      record_result "$tid" "skip" "pup $cmd_str" "$env_ctx"
      continue
    fi

    local go_schema rust_schema
    go_schema=$(python3 "$COMPARE_PY" schema < "$go_f" 2>/dev/null)
    rust_schema=$(python3 "$COMPARE_PY" schema < "$rust_f" 2>/dev/null)

    if [[ "$go_schema" == "$rust_schema" ]]; then
      record_result "$tid" "pass" "pup $cmd_str" "$env_ctx" "$go_f" "$rust_f"
    else
      # Write schemas to files for side-by-side comparison
      echo "$go_schema" > "$REPORT_DIR/go/fmt-${slug}.json"
      echo "$rust_schema" > "$REPORT_DIR/rust/fmt-${slug}.json"
      record_result "$tid" "diff" "pup $cmd_str" "$env_ctx" "$REPORT_DIR/go/fmt-${slug}.json" "$REPORT_DIR/rust/fmt-${slug}.json"
    fi
  done

  echo "  Done: ${#READ_COMMANDS[@]} tests"
}

# ── Category 6: Missing commands ─────────────────────────────────────────────
run_missing_commands_tests() {
  echo "Category 6: Missing commands detection..."
  local env_ctx="FORCE_AGENT_MODE=1"

  # Use agent schema JSON (already captured in Category 1) to extract subcommand paths
  # Also do top-level comparison
  python3 - "$REPORT_DIR" "$RESULTS_TSV" << 'PYEOF'
import json, sys, os

report_dir = sys.argv[1]
results_tsv = sys.argv[2]

def extract_commands(schema, prefix=""):
    """Recursively extract all command paths from agent schema JSON."""
    paths = set()
    commands = schema.get("commands", [])
    for cmd in commands:
        name = cmd.get("name", "")
        full = f"{prefix} {name}".strip() if prefix else name
        subs = cmd.get("subcommands", [])
        if subs:
            paths.add(full)  # group node
            for sub in subs:
                sub_name = sub.get("name", "")
                sub_full = f"{full} {sub_name}"
                paths.add(sub_full)
                # Check for deeper nesting
                for sub2 in sub.get("subcommands", []):
                    paths.add(f"{sub_full} {sub2.get('name', '')}")
        else:
            paths.add(full)
    return paths

def load_schema(filepath):
    try:
        with open(filepath) as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {}

# Top-level command comparison
go_toplevel = os.path.join(report_dir, "go", "agent-help-toplevel.json")
rust_toplevel = os.path.join(report_dir, "rust", "agent-help-toplevel.json")
go_top = load_schema(go_toplevel)
rust_top = load_schema(rust_toplevel)
go_top_cmds = {c["name"] for c in go_top.get("commands", [])}
rust_top_cmds = {c["name"] for c in rust_top.get("commands", [])}
go_only_top = sorted(go_top_cmds - rust_top_cmds)
rust_only_top = sorted(rust_top_cmds - go_top_cmds)

if go_only_top:
    go_text = "Go-only top-level commands: " + ", ".join(go_only_top)
    rust_text = "(not present in Rust)"
    go_f = os.path.join(report_dir, "go", "missing-toplevel.txt")
    rust_f = os.path.join(report_dir, "rust", "missing-toplevel.txt")
    open(go_f, "w").write(go_text)
    open(rust_f, "w").write(rust_text)
    with open(results_tsv, "a") as f:
        f.write(f"MISSING-toplevel-go-only\tdiff\tTop-level commands present in Go but not Rust\tFORCE_AGENT_MODE=1\t{go_f}\t{rust_f}\n")

if rust_only_top:
    go_text = "(not present in Go)"
    rust_text = "Rust-only top-level commands: " + ", ".join(rust_only_top)
    go_f = os.path.join(report_dir, "go", "extra-toplevel.txt")
    rust_f = os.path.join(report_dir, "rust", "extra-toplevel.txt")
    open(go_f, "w").write(go_text)
    open(rust_f, "w").write(rust_text)
    with open(results_tsv, "a") as f:
        f.write(f"MISSING-toplevel-rust-only\tdiff\tTop-level commands present in Rust but not Go\tFORCE_AGENT_MODE=1\t{go_f}\t{rust_f}\n")

if not go_only_top and not rust_only_top:
    with open(results_tsv, "a") as f:
        f.write(f"MISSING-toplevel\tpass\tAll top-level commands match\tFORCE_AGENT_MODE=1\t\t\n")

# Per-command subcommand comparison
shared = sorted(go_top_cmds & rust_top_cmds)
total = 0
for cmd in shared:
    go_f = os.path.join(report_dir, "go", f"agent-help-{cmd}.json")
    rust_f = os.path.join(report_dir, "rust", f"agent-help-{cmd}.json")
    go_schema = load_schema(go_f)
    rust_schema = load_schema(rust_f)
    go_paths = extract_commands(go_schema)
    rust_paths = extract_commands(rust_schema)
    go_only = sorted(go_paths - rust_paths)
    rust_only = sorted(rust_paths - go_paths)

    if go_only or rust_only:
        lines_go = []
        lines_rust = []
        if go_only:
            lines_go.append(f"Go-only subcommands under '{cmd}':")
            for p in go_only:
                lines_go.append(f"  pup {p}")
        if rust_only:
            lines_rust.append(f"Rust-only subcommands under '{cmd}':")
            for p in rust_only:
                lines_rust.append(f"  pup {p}")
        gf = os.path.join(report_dir, "go", f"missing-{cmd}.txt")
        rf = os.path.join(report_dir, "rust", f"missing-{cmd}.txt")
        open(gf, "w").write("\n".join(lines_go) if lines_go else "(all present)")
        open(rf, "w").write("\n".join(lines_rust) if lines_rust else "(all present)")
        with open(results_tsv, "a") as f:
            f.write(f"MISSING-{cmd}\tdiff\tpup {cmd} --help (subcommand diff)\tFORCE_AGENT_MODE=1\t{gf}\t{rf}\n")
        total += 1
    else:
        with open(results_tsv, "a") as f:
            f.write(f"MISSING-{cmd}\tpass\tpup {cmd} --help (subcommand diff)\tFORCE_AGENT_MODE=1\t\t\n")
        total += 1

print(f"  Done: {total + 1} tests")
PYEOF
}

# ── Print summary to terminal ────────────────────────────────────────────────
print_summary() {
  local pass=0 diff_count=0 error=0 skip=0 total=0
  while IFS=$'\t' read -r tid status rest; do
    total=$((total + 1))
    case "$status" in
      pass)  pass=$((pass + 1)) ;;
      diff)  diff_count=$((diff_count + 1)) ;;
      error) error=$((error + 1)) ;;
      skip)  skip=$((skip + 1)) ;;
    esac
  done < "$RESULTS_TSV"

  echo ""
  echo "=== Summary ==="
  echo "Total: $total | Pass: $pass | Diff: $diff_count | Error: $error | Skip: $skip"

  if (( diff_count > 0 )); then
    echo ""
    echo "Diffs:"
    while IFS=$'\t' read -r tid status cmd rest; do
      if [[ "$status" == "diff" ]]; then
        echo "  $tid: $cmd"
      fi
    done < <(sort "$RESULTS_TSV")
  fi

  if (( error > 0 )); then
    echo ""
    echo "Errors:"
    while IFS=$'\t' read -r tid status cmd rest; do
      if [[ "$status" == "error" ]]; then
        echo "  $tid: $cmd"
      fi
    done < <(sort "$RESULTS_TSV")
  fi
}

# ── Main ─────────────────────────────────────────────────────────────────────
main() {
  if [[ ! -x "$GO_BIN" ]]; then
    echo "ERROR: Go binary not found at $GO_BIN"
    exit 1
  fi
  if [[ ! -x "$RUST_BIN" ]]; then
    echo "Building Rust binary..."
    cargo build 2>&1 | tail -3
    if [[ ! -x "$RUST_BIN" ]]; then
      echo "ERROR: Rust binary not found at $RUST_BIN"
      exit 1
    fi
  fi

  if ! command -v dd-auth &>/dev/null; then
    echo "WARNING: dd-auth not found, skipping Category 3 (read commands)"
  fi

  local start_time=$SECONDS

  run_agent_help_tests
  run_human_help_tests

  if command -v dd-auth &>/dev/null; then
    run_read_tests
    run_format_tests
  fi

  run_auth_tests
  run_missing_commands_tests

  echo ""
  echo "Generating HTML report..."
  local report_path
  report_path=$(python3 "$COMPARE_PY" report "$RESULTS_TSV" "$REPORT_DIR" \
    "$GO_BIN" "$RUST_BIN" "$GO_VERSION" "$RUST_VERSION")
  echo "Report: $report_path"

  print_summary

  local elapsed=$(( SECONDS - start_time ))
  echo ""
  echo "Completed in ${elapsed}s"
  echo "Open: $report_path"
}

main "$@"
