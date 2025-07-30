package astinfo

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
func (manager *RpcClientManager) Generate() error {
	project := GlobalProject
	var clients map[string][]*Interface = map[string][]*Interface{}
	for _, pkg := range project.Packages {
		for _, iface := range pkg.Interfaces {
			if iface.Comment.Type == "" {
				continue
			}
			clients[iface.Comment.Type] = append(clients[iface.Comment.Type], iface)
		}
	}
	for clientType, ifaces := range clients {
		gen, ok := manager.ClientGen[clientType]
		file := createGenedFile("rpc_client_" + clientType + ".go")
		if !ok {
			continue
		}
		gen.GenerateCommon(file)
		for _, iface := range ifaces {
			err := gen.Generate(iface, file)
			if err != nil {
				return err
			}
		}
		file.save()
	}
	return nil
}

var clientGens []ClientGen

func RegisterClientGen(gen ...ClientGen) {
	clientGens = append(clientGens, gen...)
}

type ClientGen interface {
	GenerateCommon(file *GenedFile)
	Generate(iface *Interface, file *GenedFile) error
	GetName() string
}
