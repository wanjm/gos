package astinfo

import "testing"

func TestCollectRpcClientWiresInterfaceField(t *testing.T) {
	ifacePkg := &Package{Types: map[string]Typer{}}
	ifacePkg.ModPath = "example.com/prpc"
	ifacePkg.Name = "prpc"
	iface := &Interface{
		InterfaceName: "Jsinternal",
		GoSource:      &Gosourse{Pkg: ifacePkg},
	}
	iface.Comment.Type = Prpc
	ifacePkg.Types["Jsinternal"] = iface
	ifacePkg.GlobalVar = map[string]*VarField{
		"JsinternalClient": {
			Name:     "JsinternalClient",
			Type:     iface,
			GoSource: iface.GoSource,
		},
	}

	restPkg := &Package{Types: map[string]Typer{}}
	restPkg.ModPath = "example.com/rest"
	rest := &Interface{
		InterfaceName: "Upload",
		GoSource:      &Gosourse{Pkg: restPkg},
	}
	rest.Comment.Type = Restrpc
	restPkg.Types["Upload"] = rest

	bizPkg := &Package{Structs: map[string]*Struct{}}
	bizPkg.ModPath = "example.com/app/biz"
	bizPkg.Name = "biz"
	biz := &Struct{
		StructName: "CourseBiz",
		GoSource:   &Gosourse{Pkg: bizPkg},
		Fields: []*Field{
			NewSimpleField(iface, "Client"),
		},
	}
	biz.Comment.AutoGen = true
	bizPkg.Structs["CourseBiz"] = biz
	bizPkg.SortedStructNames = []string{"CourseBiz"}

	im := &InitManager{project: &MainProject{
		Packages: map[string]*Package{
			ifacePkg.ModPath: ifacePkg,
			restPkg.ModPath:  restPkg,
			bizPkg.ModPath:   bizPkg,
		},
		SortedPacakgeNames: []string{bizPkg.ModPath},
	}}
	nodes, waiting := im.collect()

	var clientNode, bizNode *DependNode
	for _, node := range nodes {
		switch node.Generator.(type) {
		case *rpcClientGenerator:
			if clientNode != nil {
				t.Fatal("restrpc must not be registered as a wire provider")
			}
			clientNode = node
		case *Struct:
			bizNode = node
		}
	}
	if clientNode == nil || bizNode == nil {
		t.Fatalf("client=%v biz=%v", clientNode, bizNode)
	}
	if got := waiting.getVariable(iface, "Client"); got != clientNode {
		t.Fatalf("interface lookup = %v, want the prpc client", got)
	}
	im.initParent(bizNode, waiting)
	if len(bizNode.Parent) != 1 || bizNode.Parent[0] != clientNode {
		t.Fatalf("CourseBiz.Client parent = %v", bizNode.Parent)
	}
	im.autofillVars = im.collectAutofillVars()
	if len(im.autofillVars) != 0 {
		t.Fatalf("without auto, autofill = %#v", im.autofillVars)
	}
	ifacePkg.GlobalVar["JsinternalClient"].Comment.AutoFill = true
	im.autofillVars = im.collectAutofillVars()
	if len(im.autofillVars) != 1 || im.autofillVars[0].Name != "JsinternalClient" {
		t.Fatalf("autofill = %#v", im.autofillVars)
	}
	im.markNeededNodes(nodes, waiting)
	if !clientNode.Needed {
		t.Fatal("@gos auto on JsinternalClient must mark the prpc client needed")
	}
}
