package app

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/kackerx/sqlmux/internal/ui"
)

// A column head's tip (§7.6「表头」), as DBeaver's: the pointer resting on a
// head shows, under it, the column's type, whether it takes NULL, its key and
// default, and the column a foreign key of it points to; it takes no key and
// no click, and goes when the pointer does. What the catalog does not say (a
// result's column) is left out: the type the result came with alone.

// headAt is the column head under the pointer in the last frame: its pane's
// table and the shown column, and its cell; nil with nothing over the grids.
func (a *App) headAt(hits []ui.Hit) (*dataTab, int, uv.Rectangle) {
	if a.overlaid() || a.cols != nil || a.drop != nil || a.dragCol != nil || a.dragTable != nil {
		return nil, 0, uv.Rectangle{}
	}
	for i := len(hits) - 1; i >= 0; i-- {
		h := hits[i]
		if !a.mouse.In(h.Rect) {
			continue
		}
		if h.Target.Kind != ui.KindHead {
			return nil, 0, uv.Rectangle{}
		}
		p := a.win().pane(h.Target.Pane)
		if p == nil {
			return nil, 0, uv.Rectangle{}
		}
		t := dataOf(p)
		if t == nil || t.props != "" || t.fieldAt(h.Target.I) < 0 {
			return nil, 0, uv.Rectangle{}
		}
		return t, h.Target.I, h.Rect
	}
	return nil, 0, uv.Rectangle{}
}

// headTip is what the tip of shown column i of t says: its name for the title,
// and its lines.
func headTip(t *dataTab, i int) (string, []string) {
	col := t.page.Cols[t.fieldAt(i)]
	cat := t.column(col.Name)
	typ := cat.Type
	if typ == "" {
		typ = col.Type
	}
	var lines []string
	line := typ
	if cat.Name != "" {
		if cat.NotNull {
			line += "  NOT NULL"
		} else {
			line += "  NULL"
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	if slices.Contains(t.cols.PK, col.Name) {
		lines = append(lines, "PRIMARY KEY")
	} else if slices.ContainsFunc(t.cols.Unique, func(u []string) bool { return slices.Contains(u, col.Name) }) {
		lines = append(lines, "UNIQUE")
	}
	if t.conn != nil && t.table.Name != "" {
		for _, k := range t.conn.keys[idOf(t.table)] {
			if k.Ref == nil {
				continue
			}
			at := slices.Index(k.Cols, col.Name)
			if at < 0 {
				continue
			}
			target := k.Ref.Table
			if k.Ref.Schema != "" && k.Ref.Schema != t.table.Schema {
				target = k.Ref.Schema + "." + target
			}
			if at < len(k.Ref.Cols) {
				target += "." + k.Ref.Cols[at]
			}
			lines = append(lines, "→ "+target)
		}
	}
	if cat.Default != "" {
		lines = append(lines, "DEFAULT "+cat.Default)
	}
	if len(cat.Enum) > 0 {
		lines = append(lines, strings.Join(cat.Enum, " | "))
	}
	return col.Name, lines
}

// drawHeadTip paints the tip of the head under the pointer, under it, or
// over it where there is no room below; the box's hit regions are dropped: a
// click goes to what is under it, as with no tip.
func (a *App) drawHeadTip(f *ui.Frame) {
	t, i, cell := a.headAt(f.Hits)
	if t == nil {
		return
	}
	title, lines := headTip(t, i)
	if len(lines) == 0 {
		return
	}
	text := strings.Join(lines, "\n")
	w := ui.Width(title) + 6
	for _, l := range lines {
		w = max(w, ui.Width(l)+4)
	}
	box, _ := ui.CompleteBox(a.window(), cell.Min, min(w, max(a.window().Dx()*2/3, 20)), len(lines), 0, ui.CompleteRows)
	n := len(f.Hits)
	ui.TextBox{Title: title, View: ui.TextView{Text: text}}.Draw(f, box)
	f.Hits = f.Hits[:n]
}

// wantHeadKeys reads the keys of the table whose head the pointer rests on,
// once, when the tree has not: its foreign keys are what the tip points to.
func (a *App) wantHeadKeys() tea.Cmd {
	t, _, _ := a.headAt(a.hits)
	if t == nil || t.conn == nil || t.table.Name == "" || t.table.Kind == "k" || !hasKeys(t.table) || t.conn.State != connReady {
		return nil
	}
	c, id := t.conn, idOf(t.table)
	if _, ok := c.keys[id]; ok || c.headAsked[id] {
		return nil
	}
	if c.headAsked == nil {
		c.headAsked = map[tableID]bool{}
	}
	c.headAsked[id] = true
	return a.fetchObjects(c, t.table)
}
