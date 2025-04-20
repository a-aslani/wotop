package remoting

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
)

// RemoteListener represents a server that listens for remote procedure calls (RPC).
//
// Fields:
//   - handler: The handler object that provides methods to be exposed via RPC.
//   - port: The port number on which the server listens for incoming connections.
type RemoteListener struct {
	handler any
	port    int
}

// RemoteCaller represents a client that makes remote procedure calls (RPC).
//
// Fields:
//   - port: The port number of the server to connect to.
type RemoteCaller struct {
	port int
}

// NewRemoteListener creates a new instance of RemoteListener.
//
// Parameters:
//   - port: The port number on which the server will listen.
//
// Returns:
//   - A pointer to a new RemoteListener instance.
func NewRemoteListener(port int) *RemoteListener {
	return &RemoteListener{
		port: port,
	}
}

// NewRemoteCaller creates a new instance of RemoteCaller.
//
// Parameters:
//   - port: The port number of the server to connect to.
//
// Returns:
//   - A pointer to a new RemoteCaller instance.
func NewRemoteCaller(port int) *RemoteCaller {
	return &RemoteCaller{port: port}
}

// SetHandler sets the handler object for the RemoteListener.
//
// The handler object must have exported methods that can be called via RPC.
//
// Parameters:
//   - Handler: The handler object to be registered for RPC.
func (r *RemoteListener) SetHandler(Handler any) {
	r.handler = Handler
}

// Run starts the RemoteListener server.
//
// This method registers the handler object for RPC, sets up an HTTP handler for RPC,
// and starts listening for incoming connections on the specified port.
//
// Logs fatal errors if the server fails to start or encounters issues.
//
// Example:
//
//	listener := NewRemoteListener(8080)
//	listener.SetHandler(&MyHandler{})
//	listener.Run()
func (r *RemoteListener) Run() {

	err := rpc.Register(r.handler)
	if err != nil {
		log.Fatal(err.Error())
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", r.port))
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("server %v is running and waiting for request from client...\n", reflect.TypeOf(r.handler))

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("serve error:", err)
		return
	}

}

// Call invokes a remote method on the server.
//
// Parameters:
//   - methodName: The name of the method to call on the server.
//   - args: The arguments to pass to the remote method.
//   - reply: A pointer to the variable where the method's response will be stored.
//
// Returns:
//   - An error if the call fails or the server is unreachable.
//
// Example:
//
//	caller := NewRemoteCaller(8080)
//	var reply string
//	err := caller.Call("MyHandler.MethodName", "argument", &reply)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *RemoteCaller) Call(methodName string, args any, reply any) error {

	var err error

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", r.port))
	if err != nil {
		return err
	}

	defer func(client *rpc.Client) {
		tempErr := client.Close()
		if err == nil {
			err = tempErr
		}
	}(client)

	err = client.Call(methodName, args, &reply)
	if err != nil {
		return err
	}

	return nil

}
