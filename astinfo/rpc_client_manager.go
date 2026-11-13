package astinfo

import (
	"sort"
	"strings"
)

type RpcClientManager struct {
	ClientGen map[string]ClientGen
}

func NewRpcClientManager() *RpcClientManager {
	return &RpcClientManager{
		ClientGen: make(map[string]ClientGen),
	}
}
func (sm *RpcClientManager) Prepare() {
	for _, callGen := range clientGens {
		sm.registerClientGen(callGen)
	}
}
func (manager *RpcClientManager) registerClientGen(gen ...ClientGen) {
	for _, gen := range gen {
		manager.ClientGen[gen.GetName()] = gen
	}
}

// Generate writes rpc client structs for every typed client interface.
// Package globals of that interface are kept only for srpc WithPrefix and restrpc assignment.
func (manager *RpcClientManager) Generate(file *GenedFile) error {
	project := GlobalProject
	var clients map[string][]*Interface = map[string][]*Interface{}
	var clientVar = make(map[*Interface]*VarField)
	pkgNames := make([]string, 0, len(project.Packages))
	for name := range project.Packages {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)
	for _, pkgName := range pkgNames {
		pkg := project.Packages[pkgName]
		varNames := make([]string, 0, len(pkg.GlobalVar))
		for name := range pkg.GlobalVar {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)
		for _, name := range varNames {
			varField := pkg.GlobalVar[name]
			iface, ok := varField.Type.(*Interface)
			if !ok || iface.Comment.Type == "" {
				continue
			}
			if _, exists := clientVar[iface]; !exists {
				clientVar[iface] = varField
			}
		}
		typeNames := make([]string, 0, len(pkg.Types))
		for name := range pkg.Types {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)
		for _, name := range typeNames {
			iface, ok := pkg.Types[name].(*Interface)
			if !ok || iface.Comment.Type == "" {
				continue
			}
			clients[iface.Comment.Type] = append(clients[iface.Comment.Type], iface)
		}
	}
	clientTypes := make([]string, 0, len(clients))
	for clientType, ifaces := range clients {
		sort.Slice(ifaces, func(i, j int) bool {
			return ifaces[i].InterfaceName < ifaces[j].InterfaceName
		})
		clientTypes = append(clientTypes, clientType)
	}
	sort.Strings(clientTypes)

	// type=prpc，srpc，restrpc；
	for _, clientType := range clientTypes {
		ifaces := clients[clientType]
		gen, ok := manager.ClientGen[clientType]
		if !ok {
			continue
		}
		file := CreateGenedFile("rpc_client_" + clientType)
		var sb strings.Builder
		gen.GenerateCommon(file)
		var rpcClientVar = make(map[*Interface]*VarField)
		for _, iface := range ifaces {
			err := gen.Generate(iface, file)
			if err != nil {
				return err
			}
			if _, ok := clientVar[iface]; ok {
				rpcClientVar[iface] = clientVar[iface]
			}
		}
		if initName := gen.InitClientVariable(rpcClientVar, file); initName != "" {
			GlobalProject.InitFuncs4All = append(GlobalProject.InitFuncs4All, initName)
		}
		file.AddBuilder(&sb)
		file.Save()
	}
	// for iface, varName := range clientVar {

	// }
	return nil
}

var clientGens []ClientGen

func RegisterClientGen(gen ...ClientGen) {
	clientGens = append(clientGens, gen...)
}

type ClientGen interface {
	GenerateCommon(file *GenedFile)
	Generate(iface *Interface, file *GenedFile) error
	// ConstructExpr returns the expression assigned to the interface variable.
	ConstructExpr(iface *Interface, file *GenedFile) string
	InitClientVariable(rpcClientVar map[*Interface]*VarField, file *GenedFile) string // 返回init函数的名字；空字符串表示没有init函数
	GetName() string                                                                  // 返回client类型的名字；
}
