package entity

type Server struct {
	ServiceName    string
	ServiceID      string
	ServiceAddress string
	ServicePort    string
}

type RegisterServer struct {
	ID      string
	Name    string
	Address string
	Port    int8
}
