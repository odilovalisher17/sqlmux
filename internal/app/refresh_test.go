package app

import (
	"slices"
	"testing"

	"github.com/kackerx/sqlmux/internal/db"
	"github.com/kackerx/sqlmux/internal/i18n"
	"github.com/kackerx/sqlmux/internal/ui"
)

// R on a connection reads it again with every database open under it; on a
// database, that one; each says it is done, as DBeaver's F5 (§7.8「刷新」).
func TestRefreshScope(t *testing.T) {
	a, c, _ := withDatabases(t)
	alpha := c.inDB("alpha")
	alpha.State, alpha.Meta, alpha.Main = connReady, c.Meta, c.Main

	a.focusPane(0)
	treeTo(t, a, "doraemon")
	cmd := a.run("tree.refresh", 0)
	if c.loading != 1 || alpha.loading != 1 || c.refreshing != 1 || alpha.refreshing != 1 {
		t.Fatalf("R on the connection: loading %d/%d, refreshing %d/%d", c.loading, alpha.loading, c.refreshing, alpha.refreshing)
	}
	answered(a, cmd)
	if c.loading != 0 || alpha.loading != 0 || c.refreshing != 0 || alpha.refreshing != 0 {
		t.Fatalf("answered: loading %d/%d, refreshing %d/%d", c.loading, alpha.loading, c.refreshing, alpha.refreshing)
	}
	if a.toast != i18n.T("已刷新 %s", alpha.label()) && a.toast != i18n.T("已刷新 %s", c.label()) {
		t.Errorf("toast %q", a.toast)
	}

	treeTo(t, a, "alpha")
	a.run("tree.refresh", 0)
	if alpha.loading != 1 || c.loading != 0 {
		t.Errorf("R on a database reads that one: %d, the server %d", alpha.loading, c.loading)
	}
}

// R on a table reads its columns again: one added shows, with its type (§7.6).
func TestGridRefreshRereadsColumns(t *testing.T) {
	a := wide(160, 45)
	tab := loadOrders(t, a, 3)
	c, id := tab.conn, idOf(tab.table)
	c.keys[id] = []db.Key{{Name: "pk", Cols: []string{"id"}}}
	if _, ok := c.cols[id]; !ok {
		t.Fatal("the page's columns are kept")
	}
	feed(t, a, "R")
	if _, ok := c.cols[id]; ok {
		t.Error("R keeps the old columns")
	}
	if _, ok := c.keys[id]; ok {
		t.Error("R keeps the old keys")
	}
}

// Refresh is in the menus of a connection, a database, a schema and a table, and
// the tree's header has a button that reads every connection again.
func TestRefreshMenuAndButton(t *testing.T) {
	a, c, _ := withDatabases(t)
	for _, text := range []string{"doraemon", "base", "public", "goal"} {
		rightClickNode(t, a, text)
		if !slices.Contains(menuLabels(a), i18n.T("刷新")) {
			t.Errorf("%s: %v", text, menuLabels(a))
		}
		a.menu = nil
	}
	btn := find(t, a, ui.Target{Kind: ui.KindButton, Pane: 0, Action: "tree.refresh.all"})
	click(a, btn.Min)
	if c.loading != 1 || c.refreshing != 1 {
		t.Errorf("the button: loading %d, refreshing %d", c.loading, c.refreshing)
	}
}
