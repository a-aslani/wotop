package wotop

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"reflect"
	"strings"
)

type ControllerStarter interface {
	Start()
}

type UsecaseRegisterer interface {
	AddUsecase(inports ...any)
	GetUsecase(nameStructType any) (any, error)
}

type ControllerRegisterer interface {
	ControllerStarter
	UsecaseRegisterer
	RegisterRouter()
	RegisterMetrics(serviceName string)
}

type RabbitmqConsumerRegisterer interface {
	UsecaseRegisterer
	Start()
	ConsumeMessage(index int, msg *amqp.Delivery)
}

type ServiceRegisterer interface {
	UsecaseRegisterer
	Start()
	//RegisterService()
}

type BaseController struct {
	inportObjs map[any]any
}

type BaseConsumer struct {
	inportObjs map[any]any
}

func NewBaseController() UsecaseRegisterer {
	return &BaseController{
		inportObjs: map[any]any{},
	}
}

func NewBaseConsumer() UsecaseRegisterer {
	return &BaseConsumer{
		inportObjs: map[any]any{},
	}
}

func NewBaseService() UsecaseRegisterer {
	return &BaseConsumer{
		inportObjs: map[any]any{},
	}
}

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

func (r *BaseController) AddUsecase(inports ...any) {
	for _, inport := range inports {
		x := reflect.ValueOf(inport).Elem().Type().String()
		packagePath := x[:strings.Index(x, ".")]
		r.inportObjs[packagePath] = inport
	}
}

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

func (b BaseConsumer) AddUsecase(inports ...any) {
	for _, inport := range inports {
		x := reflect.ValueOf(inport).Elem().Type().String()
		packagePath := x[:strings.Index(x, ".")]
		b.inportObjs[packagePath] = inport
	}
}
