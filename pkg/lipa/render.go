package lipa

import (
	"bytes"
	"encoding/json"
	"html/template"
)

func renderHTML(root *Node, options Options) string {
	payload, _ := json.Marshal(root)
	data := struct {
		Title       string
		ExpandDepth int
		ShowHidden  bool
		TreeJSON    template.JS
	}{
		Title:       options.Title,
		ExpandDepth: options.ExpandDepth,
		ShowHidden:  options.ShowHidden,
		TreeJSON:    template.JS(payload),
	}

	var buf bytes.Buffer
	_ = pageTemplate.Execute(&buf, data)
	return buf.String()
}

var pageTemplate = template.Must(template.New("lipa").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root {
  color-scheme: light;
  --bg: #ffffff;
  --panel: #ffffff;
  --line: #d9dee8;
  --text: #1f2937;
  --muted: #6b7280;
  --soft: #f6f7f9;
  --border: #e5e7eb;
  --blue: #2563eb;
  --blue-soft: #eff6ff;
  --green: #047857;
  --red: #dc2626;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100vh;
  overflow: hidden;
  background: var(--bg);
  color: var(--text);
  font: 13px/1.45 ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
button, input {
  font: inherit;
}
button {
  border: 1px solid var(--border);
  border-radius: 6px;
  background: #fff;
  color: var(--text);
  padding: 6px 8px;
  cursor: pointer;
}
button:hover {
  border-color: #bfd1ff;
  color: var(--blue);
  background: #fafcff;
}
#viewport {
  width: 100vw;
  height: 100vh;
  cursor: grab;
  touch-action: none;
}
#viewport.dragging { cursor: grabbing; }
#canvas {
  position: absolute;
  top: 34px;
  left: 34px;
  transform-origin: 0 0;
}
.tree, .tree ul {
  margin: 0;
  padding: 0 0 0 22px;
  list-style: none;
}
.tree {
  min-width: 720px;
  padding-bottom: 80px;
}
.tree li {
  position: relative;
  margin: 5px 0;
}
.tree li::before {
  content: "";
  position: absolute;
  top: 17px;
  left: -13px;
  width: 12px;
  border-top: 1px solid var(--line);
}
.tree ul::before {
  content: "";
  position: absolute;
  top: 31px;
  bottom: 12px;
  left: 8px;
  border-left: 1px solid var(--line);
}
.node {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  max-width: 1040px;
  min-height: 30px;
  padding: 6px 8px;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: var(--panel);
  box-shadow: 0 1px 2px rgba(15, 23, 42, .04);
  white-space: nowrap;
  user-select: none;
}
.node:hover {
  border-color: #c7d2fe;
  box-shadow: 0 3px 12px rgba(37, 99, 235, .08);
}
.node.selected {
  border-color: var(--blue);
  background: var(--blue-soft);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, .10);
}
.node.match:not(.selected) {
  border-color: #93c5fd;
  background: #f8fbff;
}
.toggle {
  width: 20px;
  height: 20px;
  display: inline-grid;
  place-items: center;
  padding: 0;
  border-color: #d7dce5;
  background: #f9fafb;
  color: var(--blue);
  font-weight: 700;
}
.toggle.empty { visibility: hidden; }
.name {
  color: #111827;
  font-weight: 700;
}
.badge {
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--soft);
  color: var(--muted);
  font-size: 11px;
}
.type {
  color: #4b5563;
}
.value {
  color: var(--green);
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 560px;
}
.nil { color: var(--muted); }
.cycle, .trunc { color: var(--red); }
.collapsed > ul { display: none; }
.hide-hidden .hidden-node { display: none; }
#panel {
  position: fixed;
  top: 14px;
  right: 14px;
  z-index: 3;
  width: 310px;
  max-height: calc(100vh - 28px);
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: rgba(255, 255, 255, .96);
  box-shadow: 0 10px 28px rgba(15, 23, 42, .10);
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.panel-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 700;
}
.panel-meta {
  color: var(--muted);
  font-size: 12px;
}
.search {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 6px;
}
.search input {
  min-width: 0;
  border: 1px solid var(--border);
  border-radius: 7px;
  padding: 7px 9px;
  outline: none;
}
.search input:focus {
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, .10);
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.check {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--muted);
}
#selection {
  min-height: 36px;
  padding: 8px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: #fafafa;
  color: var(--muted);
}
#selection strong {
  color: var(--text);
}
.source-snippet {
  margin-top: 8px;
  border-top: 1px solid var(--border);
  padding-top: 8px;
}
.snippet-head {
  margin-bottom: 5px;
  color: var(--muted);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.source-snippet pre {
  margin: 0;
  max-width: 100%;
  overflow: auto;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: #fff;
  padding: 8px;
  color: #374151;
  font: 12px/1.45 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}
.snippet-caret {
  color: var(--blue);
  font-weight: 700;
}
#navigator {
  overflow: auto;
  min-height: 160px;
  max-height: 38vh;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: #fff;
}
.nav-item {
  display: block;
  width: 100%;
  border: 0;
  border-radius: 0;
  padding: 7px 9px;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nav-item:hover {
  background: #f8fbff;
}
.nav-item.active {
  background: var(--blue-soft);
  color: var(--blue);
}
.hidden { display: none; }
</style>
</head>
<body>
<div id="viewport"><div id="canvas"><ul class="tree" id="tree"></ul></div></div>
<aside id="panel">
  <div class="panel-head">
    <div class="panel-title">{{.Title}}</div>
    <button id="hidePanel" title="Hide panel">×</button>
  </div>
  <div class="search">
    <input id="search" placeholder="Search node..." autocomplete="off">
    <button id="prev" title="Previous match">↑</button>
    <button id="next" title="Next match">↓</button>
  </div>
  <div class="panel-meta" id="resultMeta">0 nodes</div>
  <label class="check"><input id="showHiddenToggle" type="checkbox" {{if .ShowHidden}}checked{{end}}> Show hidden</label>
  <div id="selection">No node selected</div>
  <div class="grid">
    <button id="focus">Focus</button>
    <button id="fit">Reset view</button>
    <button id="expandNode">Expand node</button>
    <button id="collapseNode">Collapse node</button>
    <button id="expand">Expand all</button>
    <button id="collapse">Collapse all</button>
  </div>
  <div id="navigator"></div>
</aside>
<button id="showPanel" class="hidden" style="position:fixed;right:14px;top:14px;z-index:4;">Menu</button>
<script>
const root = {{.TreeJSON}};
const initialDepth = {{.ExpandDepth}};
let showHidden = {{if .ShowHidden}}true{{else}}false{{end}};
const tree = document.getElementById('tree');
const viewport = document.getElementById('viewport');
const canvas = document.getElementById('canvas');
const panel = document.getElementById('panel');
const showPanel = document.getElementById('showPanel');
const searchInput = document.getElementById('search');
const navigatorBox = document.getElementById('navigator');
const resultMeta = document.getElementById('resultMeta');
const selectionBox = document.getElementById('selection');
const showHiddenToggle = document.getElementById('showHiddenToggle');
let scale = 1;
let x = 34;
let y = 34;
let dragging = false;
let lastX = 0;
let lastY = 0;
let selected = null;
let matches = [];
let matchIndex = -1;
const entries = [];

function esc(text) {
  return String(text ?? '').replace(/[&<>"']/g, ch => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));
}

function labelFor(node) {
  return [node.name, node.kind, node.type, node.value, node.ref].filter(Boolean).join(' ');
}

function snippetLocation(snippet) {
  let location = snippet.fileName || '';
  if (snippet.line) location += (location ? ':' : 'line ') + snippet.line;
  if (snippet.column) location += ':' + snippet.column;
  return location || 'source';
}

function snippetHTML(snippet) {
  if (!snippet) return '';
  return '<div class="source-snippet">' +
    '<div class="snippet-head">' + esc(snippetLocation(snippet)) + '</div>' +
    '<pre><code>' + esc(snippet.text || '') + '</code>\n<span class="snippet-caret">' + esc(snippet.marker || '^') + '</span></pre>' +
    '</div>';
}

function renderNode(node, depth = 0) {
  const li = document.createElement('li');
  const hasChildren = node.children && node.children.length;
  if (node.hidden) li.classList.add('hidden-node');
  if (hasChildren && depth >= initialDepth) li.classList.add('collapsed');

  const row = document.createElement('div');
  row.className = 'node';
  row.dataset.nodeId = node.id;
  row.innerHTML =
    '<button class="toggle ' + (hasChildren ? '' : 'empty') + '">' + (hasChildren && depth >= initialDepth ? '+' : '-') + '</button>' +
    '<span class="name">' + esc(node.name) + '</span>' +
    '<span class="badge">' + esc(node.kind) + '</span>' +
    '<span class="type">' + esc(node.type) + '</span>' +
    (node.value ? '<span class="value ' + (node.nil ? 'nil ' : '') + (node.cycle ? 'cycle ' : '') + (node.trunc ? 'trunc ' : '') + '">' + esc(node.value) + '</span>' : '') +
    (node.ref ? '<span class="badge">' + esc(node.ref) + '</span>' : '');
  li.appendChild(row);

  const entry = {node, li, row, depth, label: labelFor(node).toLowerCase()};
  entries.push(entry);

  row.addEventListener('click', event => {
    if (event.target.classList.contains('toggle')) return;
    toggleEntry(entry);
    selectEntry(entry, false);
  });

  const toggle = row.querySelector('.toggle');
  toggle.addEventListener('click', event => {
    event.stopPropagation();
    toggleEntry(entry);
    selectEntry(entry, false);
  });

  if (hasChildren) {
    const ul = document.createElement('ul');
    for (const child of node.children) ul.appendChild(renderNode(child, depth + 1));
    li.appendChild(ul);
  }
  return li;
}

function toggleEntry(entry) {
  if (!entry || !entry.li.querySelector(':scope > ul')) return;
  const collapsed = !entry.li.classList.contains('collapsed');
  setEntryCollapsed(entry, collapsed);
}

function setEntryCollapsed(entry, collapsed) {
  if (!entry || !entry.li.querySelector(':scope > ul')) return;
  entry.li.classList.toggle('collapsed', collapsed);
  entry.row.querySelector('.toggle').textContent = collapsed ? '+' : '-';
}

function setCollapsed(collapsed) {
  for (const entry of entries) setEntryCollapsed(entry, collapsed);
}

function reveal(entry) {
  let li = entry && entry.li ? entry.li.parentElement.closest('li') : null;
  while (li) {
    const parentEntry = entries.find(item => item.li === li);
    if (parentEntry) setEntryCollapsed(parentEntry, false);
    li = li.parentElement.closest('li');
  }
}

function selectEntry(entry, center) {
  if (!entry) return;
  if (selected) selected.row.classList.remove('selected');
  selected = entry;
  selected.row.classList.add('selected');
  reveal(selected);
  selectionBox.innerHTML =
    '<strong>' + esc(selected.node.name) + '</strong><br>' +
    esc(selected.node.type || selected.node.kind) +
    '<br><span class="panel-meta">click node to fold/unfold</span>' +
    snippetHTML(selected.node.snippet);
  updateNavigatorActive();
  if (center) centerEntry(selected);
}

function centerEntry(entry) {
  if (!entry) return;
  reveal(entry);
  requestAnimationFrame(() => {
    const rect = entry.row.getBoundingClientRect();
    x += (viewport.clientWidth / 2) - (rect.left + rect.width / 2);
    y += (viewport.clientHeight / 2) - (rect.top + rect.height / 2);
    updateTransform();
  });
}

function updateTransform() {
  canvas.style.transform = 'translate(' + x + 'px, ' + y + 'px) scale(' + scale + ')';
}

function resetView() {
  scale = 1;
  x = 34;
  y = 34;
  updateTransform();
}

function applySearch() {
  const query = searchInput.value.trim().toLowerCase();
  matches = [];
  matchIndex = -1;
  for (const entry of entries) {
    const visible = showHidden || !entry.node.hidden;
    const ok = visible && query !== '' && entry.label.includes(query);
    entry.row.classList.toggle('match', ok);
    if (ok) matches.push(entry);
  }
  const visibleEntries = entries.filter(entry => showHidden || !entry.node.hidden);
  renderNavigator(query ? matches : visibleEntries);
  resultMeta.textContent = query ? (matches.length + ' matches') : (visibleEntries.length + ' nodes');
  if (matches.length > 0) {
    matchIndex = 0;
    selectEntry(matches[0], true);
  }
}

function moveMatch(delta) {
  if (matches.length === 0) return;
  matchIndex = (matchIndex + delta + matches.length) % matches.length;
  selectEntry(matches[matchIndex], true);
}

function renderNavigator(items) {
  navigatorBox.innerHTML = '';
  const limit = Math.min(items.length, 300);
  for (let i = 0; i < limit; i++) {
    const entry = items[i];
    const item = document.createElement('button');
    item.className = 'nav-item';
    item.dataset.nodeId = entry.node.id;
    item.style.paddingLeft = Math.min(18 + entry.depth * 10, 80) + 'px';
    item.textContent = entry.node.name + ' · ' + (entry.node.type || entry.node.kind);
    item.addEventListener('click', () => selectEntry(entry, true));
    navigatorBox.appendChild(item);
  }
  if (items.length > limit) {
    const rest = document.createElement('div');
    rest.className = 'panel-meta';
    rest.style.padding = '8px 9px';
    rest.textContent = '+' + (items.length - limit) + ' more, use search';
    navigatorBox.appendChild(rest);
  }
  updateNavigatorActive();
}

function updateNavigatorActive() {
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.toggle('active', selected && item.dataset.nodeId === selected.node.id);
  });
}

tree.appendChild(renderNode(root));
tree.classList.toggle('hide-hidden', !showHidden);
renderNavigator(entries.filter(entry => showHidden || !entry.node.hidden));
resultMeta.textContent = entries.filter(entry => showHidden || !entry.node.hidden).length + ' nodes';
updateTransform();
selectEntry(entries[0], false);

viewport.addEventListener('pointerdown', event => {
  if (event.target.closest('#panel') || event.target.closest('#showPanel')) return;
  dragging = true;
  lastX = event.clientX;
  lastY = event.clientY;
  viewport.classList.add('dragging');
  viewport.setPointerCapture(event.pointerId);
});
viewport.addEventListener('pointermove', event => {
  if (!dragging) return;
  x += event.clientX - lastX;
  y += event.clientY - lastY;
  lastX = event.clientX;
  lastY = event.clientY;
  updateTransform();
});
viewport.addEventListener('pointerup', event => {
  dragging = false;
  viewport.classList.remove('dragging');
  viewport.releasePointerCapture(event.pointerId);
});
viewport.addEventListener('wheel', event => {
  event.preventDefault();
  const before = scale;
  const factor = event.deltaY < 0 ? 1.1 : 0.9;
  scale = Math.min(3, Math.max(.18, scale * factor));
  const rect = viewport.getBoundingClientRect();
  const mx = event.clientX - rect.left;
  const my = event.clientY - rect.top;
  x = mx - (mx - x) * (scale / before);
  y = my - (my - y) * (scale / before);
  updateTransform();
}, {passive: false});

searchInput.addEventListener('input', applySearch);
searchInput.addEventListener('keydown', event => {
  if (event.key === 'Enter') moveMatch(event.shiftKey ? -1 : 1);
});
document.getElementById('prev').addEventListener('click', () => moveMatch(-1));
document.getElementById('next').addEventListener('click', () => moveMatch(1));
document.getElementById('focus').addEventListener('click', () => centerEntry(selected));
document.getElementById('fit').addEventListener('click', resetView);
document.getElementById('expand').addEventListener('click', () => setCollapsed(false));
document.getElementById('collapse').addEventListener('click', () => setCollapsed(true));
document.getElementById('expandNode').addEventListener('click', () => setEntryCollapsed(selected, false));
document.getElementById('collapseNode').addEventListener('click', () => setEntryCollapsed(selected, true));
showHiddenToggle.addEventListener('change', () => {
  showHidden = showHiddenToggle.checked;
  tree.classList.toggle('hide-hidden', !showHidden);
  applySearch();
});
document.getElementById('hidePanel').addEventListener('click', () => {
  panel.classList.add('hidden');
  showPanel.classList.remove('hidden');
});
showPanel.addEventListener('click', () => {
  panel.classList.remove('hidden');
  showPanel.classList.add('hidden');
  searchInput.focus();
});
document.addEventListener('keydown', event => {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'f') {
    event.preventDefault();
    panel.classList.remove('hidden');
    showPanel.classList.add('hidden');
    searchInput.focus();
    searchInput.select();
  }
  if (event.key === 'Escape') {
    searchInput.value = '';
    applySearch();
  }
});
</script>
</body>
</html>`))
