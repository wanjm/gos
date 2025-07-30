package astinfo

import "strings"

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

// Generate
func (manager *RpcClientManager) Generate(file *GenedFile) error {
	project := GlobalProject
	var clients map[string][]*Interface = map[string][]*Interface{}
	var clientVar = make(map[*Interface]*VarField)
	for _, pkg := range project.Packages {
		for _, iface := range pkg.Interfaces {
			if iface.Comment.Type == "" {
				continue
			}
			clients[iface.Comment.Type] = append(clients[iface.Comment.Type], iface)
		}
		for _, varField := range pkg.GlobalVar {
			if iface, ok := varField.Type.(*Interface); ok {
				if iface.Comment.Type != "" {
					clientVar[iface] = varField
				}
			}
		}
	}
	for clientType, ifaces := range clients {
		gen, ok := manager.ClientGen[clientType]
		if !ok {
			continue
		}
		file := createGenedFile("rpc_client_" + clientType + ".go")
		var sb strings.Builder
		gen.GenerateCommon(file)
		var rpcClientVar = make(map[*Interface]*VarField)
		for _, iface := range ifaces {
			err := gen.Generate(iface, file)
			if err != nil {
				return err
			}
			rpcClientVar[iface] = clientVar[iface]
		}
		gen.InitClientVariable(rpcClientVar, file)
		file.AddBuilder(&sb)
		file.save()
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
	InitClientVariable(rpcClientVar map[*Interface]*VarField, file *GenedFile) string // 返回init函数的名字；
	GetName() string                                                                  // 返回client类型的名字；
}
