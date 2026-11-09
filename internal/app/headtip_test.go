package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/kackerx/sqlmux/internal/db"
	"github.com/kackerx/sqlmux/internal/ui"
)

// The pointer on a column head shows its type, NULL or not, its key and
// default, and where a foreign key of it points; off the heads, no tip; a
// click on a head still sorts, the tip taking no click (§7.6「表头」).
func TestHeadTip(t *testing.T) {
	a := wide(160, 45)
	tab := loadOrders(t, a, 5)
	c := tab.conn
	c.keys[idOf(tab.table)] = []db.Key{
		{Name: "t_order_pkey", Cols: []string{"id"}},
		{Name: "fk_status", Cols: []string{"status"}, Ref: &db.Ref{Schema: "ref", Table: "statuses", Cols: []string{"code"}}},
	}
	hover := func(i int) string {
		r := find(t, a, ui.Target{Kind: ui.KindHead, Pane: a.focused().ID, I: i})
		a.Update(tea.MouseMotionMsg{X: r.Min.X + 1, Y: r.Min.Y})
		return a.render().String()
	}
	if s := hover(0); !strings.Contains(s, "bigint  NOT NULL") || !strings.Contains(s, "PRIMARY KEY") {
		t.Errorf("id's tip:\n%s", s)
	}
	if s := hover(1); !strings.Contains(s, "order_status  NULL") || !strings.Contains(s, "→ ref.statuses.code") || !strings.Contains(s, "pending | running") {
		t.Errorf("status's tip:\n%s", s)
	}
	a.Update(tea.MouseMotionMsg{X: 100, Y: 30}) // over the rows
	if s := a.render().String(); strings.Contains(s, "NOT NULL") || strings.Contains(s, "→ ref.statuses") {
		t.Error("the tip stays off the heads")
	}
	r := find(t, a, ui.Target{Kind: ui.KindHead, Pane: a.focused().ID, I: 2})
	a.Update(tea.MouseMotionMsg{X: r.Min.X + 1, Y: r.Min.Y})
	a.View()
	click(a, uv.Pos(r.Min.X+1, r.Min.Y))
	if len(tab.request.order) == 0 || tab.request.order[0].Col != "amount" {
		t.Errorf("a click on the head sorts: %+v", tab.request.order)
	}
}

// The keys of a table whose head is hovered are read once, when the tree has
// not read them.
func TestHeadTipAsksKeys(t *testing.T) {
	a := wide(160, 45)
	tab := loadOrders(t, a, 5)
	r := find(t, a, ui.Target{Kind: ui.KindHead, Pane: a.focused().ID, I: 0})
	_, cmd := a.Update(tea.MouseMotionMsg{X: r.Min.X + 1, Y: r.Min.Y})
	if cmd == nil || !tab.conn.headAsked[idOf(tab.table)] {
		t.Fatal("hovering a head asks for the table's keys")
	}
	a.Update(tea.MouseMotionMsg{X: r.Min.X + 2, Y: r.Min.Y})
	if len(tab.conn.headAsked) != 1 {
		t.Error("asked once")
	}
}
