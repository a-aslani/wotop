package wotop

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"reflect"
	"strings"
)

// ControllerStarter defines an interface for starting a controller.
type ControllerStarter interface {
	// Start initializes and starts the controller.
	Start()
}

// UsecaseRegisterer defines an interface for registering and retrieving use cases.
type UsecaseRegisterer interface {
	// AddUsecase registers one or more use cases.
	//
	// Parameters:
	//   - inports: Variadic parameter representing the use cases to be registered.
	AddUsecase(inports ...any)

	// GetUsecase retrieves a registered use case by its type.
	//
	// Parameters:
	//   - nameStructType: The type of the use case to retrieve.
	//
	// Returns:
	//   - The registered use case, or an error if it is not found.
	GetUsecase(nameStructType any) (any, error)
}

// ControllerRegisterer defines an interface that combines controller starting,
// use case registration, and additional functionalities like router and metrics registration.
type ControllerRegisterer interface {
	ControllerStarter
	UsecaseRegisterer

	// RegisterRouter sets up the router for the controller.
	RegisterRouter()

	// RegisterMetrics sets up metrics for the service.
	//
	// Parameters:
	//   - serviceName: The name of the service for which metrics are being registered.
	RegisterMetrics(serviceName string)
}

// RabbitmqConsumerRegisterer defines an interface for registering and consuming RabbitMQ messages.
type RabbitmqConsumerRegisterer interface {
	UsecaseRegisterer

	// Start initializes and starts the RabbitMQ consumer.
	Start()

	// ConsumeMessage processes a RabbitMQ message.
	//
	// Parameters:
	//   - index: The index of the message.
	//   - msg: The RabbitMQ message to be consumed.
	ConsumeMessage(index int, msg *amqp.Delivery)
}

// ServiceRegisterer defines an interface for registering and starting services.
type ServiceRegisterer interface {
	UsecaseRegisterer

	// Start initializes and starts the service.
	Start()
}

// BaseController provides a base implementation for use case registration.
type BaseController struct {
	inportObjs map[any]any // A map to store registered use cases.
}

// BaseConsumer provides a base implementation for use case registration in consumers.
type BaseConsumer struct {
	inportObjs map[any]any // A map to store registered use cases.
}

// NewBaseController creates a new instance of BaseController.
//
// Returns:
//   - A UsecaseRegisterer instance for registering use cases.
func NewBaseController() UsecaseRegisterer {
	return &BaseController{
		inportObjs: map[any]any{},
	}
}

// NewBaseConsumer creates a new instance of BaseConsumer.
//
// Returns:
//   - A UsecaseRegisterer instance for registering use cases.
func NewBaseConsumer() UsecaseRegisterer {
	return &BaseConsumer{
		inportObjs: map[any]any{},
	}
}

// NewBaseService creates a new instance of BaseConsumer for services.
//
// Returns:
//   - A UsecaseRegisterer instance for registering use cases.
func NewBaseService() UsecaseRegisterer {
	return &BaseConsumer{
		inportObjs: map[any]any{},
	}
}

// GetUsecase retrieves a registered use case by its type.
//
// Parameters:
//   - nameStructType: The type of the use case to retrieve.
//
// Returns:
//   - The registered use case, or an error if it is not found.
func (r *BaseController) GetUsecase(nameStructType any) (any, error) {
	x := reflect.TypeOf(nameStructType).String()
	packageName := x[:strings.Index(x, ".")]
	uc, ok := r.inportObjs[packageName]
	if !ok {
		msg := "usecase with package \"%s\" is not registered yet in application"
		return nil, fmt.Errorf(msg, packageName)
	}
	return uc, nil
}

// AddUsecase registers one or more use cases.
//
// Parameters:
//   - inports: Variadic parameter representing the use cases to be registered.
func (r *BaseController) AddUsecase(inports ...any) {
	for _, inport := range inports {
		x := reflect.ValueOf(inport).Elem().Type().String()
		packagePath := x[:strings.Index(x, ".")]
		r.inportObjs[packagePath] = inport
	}
}

// GetUsecase retrieves a registered use case by its type.
//
// Parameters:
//   - nameStructType: The type of the use case to retrieve.
//
// Returns:
//   - The registered use case, or an error if it is not found.
func (b BaseConsumer) GetUsecase(nameStructType any) (any, error) {
	x := reflect.TypeOf(nameStructType).String()
	packageName := x[:strings.Index(x, ".")]
	uc, ok := b.inportObjs[packageName]
	if !ok {
		msg := "usecase with package \"%s\" is not registered yet in application"
		return nil, fmt.Errorf(msg, packageName)
	}
	return uc, nil
}

// AddUsecase registers one or more use cases.
//
// Parameters:
//   - inports: Variadic parameter representing the use cases to be registered.
func (b BaseConsumer) AddUsecase(inports ...any) {
	for _, inport := range inports {
		x := reflect.ValueOf(inport).Elem().Type().String()
		packagePath := x[:strings.Index(x, ".")]
		b.inportObjs[packagePath] = inport
	}
}
