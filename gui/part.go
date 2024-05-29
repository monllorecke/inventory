package gui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mbertschler/blocks/html"
	"github.com/mbertschler/inventory/parts"
	"github.com/mbertschler/inventory/lib/guiapi"
)

func init() {
	guiapi.DefaultHandler.Functions["editPart"] = editPartAction
	guiapi.DefaultHandler.Functions["savePart"] = savePartAction
	guiapi.DefaultHandler.Functions["deletePart"] = deletePartAction
	guiapi.DefaultHandler.Functions["prepareTruck"] = prepareTruckAction
	guiapi.DefaultHandler.Functions["ship"] = shipAction
}

func partPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/part/")
	part, err := parts.ByID(id)
	if err != nil {
		log.Println(err)
	}
	page := mainLayout(viewPartBlock(part))
	err = html.Render(w, page)
	if err != nil {
		log.Println(err)
	}
}

func editPartAction(args json.RawMessage) (*guiapi.Result, error) {
	var id string
	err := json.Unmarshal(args, &id)
	if err != nil {
		return nil, err
	}
	part, err := parts.ByID(id)
	if err != nil {
		return nil, err
	}
	return guiapi.Replace("#container", editPartBlock(part, ""))
}

func savePartAction(args json.RawMessage) (*guiapi.Result, error) {
	type input struct {
		ID         string
		New        string
		Reference  string
		Weight     string
		Quantity   string
		Location   string
		Supplier   string
		Dimensions string
		Status     string
		ArrivalDate string
	}
	var in input
	err := json.Unmarshal(args, &in)
	if err != nil {
		return nil, err
	}
	var p *parts.Part
	if in.New == "true" {
		p, err = parts.Create()
	} else {
		p, err = parts.ByID(in.ID)
	}
	if err != nil {
		return nil, err
	}
	p.Reference = in.Reference
	weight, err := strconv.ParseFloat(in.Weight, 64)
	if err != nil {
		return nil, err
	}
	p.Weight = weight
	quant, err := strconv.Atoi(in.Quantity)
	if err != nil {
		return nil, err
	}
	p.Quantity = quant
	p.Location = in.Location
	p.Supplier = in.Supplier
	p.Dimensions = in.Dimensions
	p.Status = in.Status
	p.ArrivalDate, err = time.Parse(time.RFC3339, in.ArrivalDate)
	if err != nil {
		return nil, err
	}

	// store part
	err = parts.Store(p)
	if err != nil {
		return nil, err
	}

	return guiapi.Replace("#container", viewPartBlock(p))
}

func deletePartAction(args json.RawMessage) (*guiapi.Result, error) {
	var id string
	err := json.Unmarshal(args, &id)
	if err != nil {
		return nil, err
	}
	err = parts.DeleteByID(id)
	if err != nil {
		return nil, err
	}
	return guiapi.Redirect("/")
}

func prepareTruckAction(args json.RawMessage) (*guiapi.Result, error) {
	var id string
	err := json.Unmarshal(args, &id)
	if err != nil {
		return nil, err
	}
	part, err := parts.ByID(id)
	if err != nil {
		return nil, err
	}
	part.Status = "en preparación de camión"
	err = parts.Store(part)
	if err != nil {
		return nil, err
	}
	return guiapi.Redirect("/")
}

func shipAction(args json.RawMessage) (*guiapi.Result, error) {
	var id string
	err := json.Unmarshal(args, &id)
	if err != nil {
		return nil, err
	}
	part, err := parts.ByID(id)
	if err != nil {
		return nil, err
	}
	part.Status = "en salida de camión"
	err = parts.Store(part)
	if err != nil {
		return nil, err
	}
	return guiapi.Redirect("/")
}

func viewPartBlock(p *parts.Part) html.Block {
	editAction := fmt.Sprintf("guiapi('editPart', '%s')", p.ID())
	deleteAction := fmt.Sprintf("guiapi('deletePart', '%s')", p.ID())
	prepareTruckAction := fmt.Sprintf("guiapi('prepareTruck', '%s')", p.ID())
	shipAction := fmt.Sprintf("guiapi('ship', '%s')", p.ID())

	var rows html.Blocks
	r := func(k, v string) html.Block {
		return html.Elem("tr", nil,
			html.Elem("td", nil, html.Text(k)),
			html.Elem("td", nil, html.Text(v)),
		)
	}
	rows.Add(r("Reference", p.Reference))
	rows.Add(r("Weight", fmt.Sprintf("%.2f", p.Weight)))
	rows.Add(r("Quantity", fmt.Sprint(p.Quantity)))
	rows.Add(r("Location", p.Location))
	rows.Add(r("Supplier", p.Supplier))
	rows.Add(r("Dimensions", p.Dimensions))
	rows.Add(r("Status", p.Status))
	rows.Add(r("Arrival Date", p.ArrivalDate.Format(time.RFC3339)))

	return html.Div(nil,
		html.Div(nil,
			html.A(html.Href("/"),
				html.Button(html.Class("ui button"),
					html.Text("< List"),
				),
			),
			html.Button(html.Class("ui button").
				Attr("onclick", editAction),
				html.Text("Edit"),
			),
			html.Button(html.Class("ui red button").
				Attr("onclick", deleteAction),
				html.Text("Delete"),
			),
			html.A(html.Href(fmt.Sprintf("/part/generate-qr?id=%s", p.ID())).Attr("download", "true"),
				html.Button(html.Class("ui green button"),
					html.Text("Download QR Code"),
				),
			),
			html.Button(html.Class("ui yellow button").
				Attr("onclick", prepareTruckAction),
				html.Text("Prepare Truck"),
			),
			html.Button(html.Class("ui blue button").
				Attr("onclick", shipAction),
				html.Text("Ship"),
			),
		),
		html.H1(nil, html.Text(p.Reference)),
		html.Elem("table", html.Class("ui celled table"),
			html.Elem("tbody", nil,
				rows,
			),
		),
	)
}

func editPartBlock(p *parts.Part, code string) html.Block {
	isNew := false
	if p == nil {
		isNew = true
		p = &parts.Part{}
	}
	if code != "" {
		p.Reference = code
	}
	cancelAction := fmt.Sprintf("guiapi('viewPart', '%s')", p.ID())
	saveAction := "sendForm('savePart', '.ga-edit-part')"
	return html.Div(nil,
		html.Div(nil,
			html.Button(html.Class("ui button").
				Attr("onclick", cancelAction),
				html.Text("Cancel"),
			),
			html.Button(html.Class("ui green button").
				Attr("onclick", saveAction),
				html.Text("Save"),
			),
		),
		html.Div(html.Class("ui form"),
			html.Input(html.Type("hidden").Name("New").Value(isNew).Class("ga-edit-part")),
			html.Input(html.Type("hidden").Name("ID").Value(p.ID()).Class("ga-edit-part")),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Reference")),
				html.Input(html.Type("Text").Name("Reference").Value(p.Reference).Class("ga-edit-part")),
			),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Weight")),
				html.Input(html.Type("Number").Name("Weight").Value(fmt.Sprintf("%.2f", p.Weight)).Class("ga-edit-part")),
			),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Quantity")),
				html.Input(html.Type("Number").Name("Quantity").Value(fmt.Sprint(p.Quantity)).Class("ga-edit-part")),
			),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Location")),
				html.Input(html.Type("Text").Name("Location").Value(p.Location).Class("ga-edit-part")),
			),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Supplier")),
				html.Input(html.Type("Text").Name("Supplier").Value(p.Supplier).Class("ga-edit-part")),
			),
			html.Div(html.Class("field"),
				html.Label(nil, html.Text("Dimensions")),
				html.Input(html.Type("Text").Name("Dimensions").Value(p.Dimensions).Class("ga-edit-part")),
			),
		),
	)
}
